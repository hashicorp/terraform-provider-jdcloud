package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	"time"
)

//---------------------------------------------------------------------------------	ATTACHMENT-SCHEMA-HELPERS

// This function will return the latest status of a disk level 1 -> Based on QueryInstanceDetail
// *We've already had a disk status refresher, this one is generated since by checking the
// attachment status we are going to describe VMs rather than disk itself
func diskAttachmentStatusRefreshFunc(d *schema.ResourceData, meta interface{}, instanceId, diskId string) resource.StateRefreshFunc {
	return func() (diskItem interface{}, diskState string, e error) {

		resp, err := QueryInstanceDetail(d, meta, instanceId)
		if err != nil {
			return nil, "", err
		}

		// We've found the expected disk
		for _, d := range resp.Result.Instance.DataDisks {
			if d.CloudDisk.DiskId == diskId {
				return d, d.Status, nil
			}
		}

		// We have not found the desired one
		return nil, "DiskNotFound", fmt.Errorf("DiskNotFound")
	}
}

// This function will send a request of attachment,do not wait,just send, level 0
func performDiskAttach(d *schema.ResourceData, meta interface{}, req *apis.AttachDiskRequest) (requestId string, e error) {

	config := meta.(*JDCloudConfig)
	c := client.NewVmClient(config.Credential)
	e = resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := c.AttachDisk(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			requestId = resp.RequestID
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return requestId, e
}

// This function will send a request of detachment,do not wait,just send, level 0
func performDiskDetach(d *schema.ResourceData, meta interface{}, req *apis.DetachDiskRequest) error {

	config := meta.(*JDCloudConfig)
	c := client.NewVmClient(config.Credential)
	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := c.DetachDisk(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

// This function will wait until certain status has been reached, level 2 -> based on diskAttachmentStatusRefreshFunc
func diskAttachmentWaiter(d *schema.ResourceData, meta interface{}, instanceId, diskId string, pending, target []string) (err error) {

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    diskAttachmentStatusRefreshFunc(d, meta, instanceId, diskId),
		Delay:      3 * time.Second,
		Timeout:    2 * time.Minute,
		MinTimeout: 1 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in AttachingDisk/WaitingDiskAttaching ,err message:%v", err)
	}
	return nil
}

//---------------------------------------------------------------------------------	ATTACHMENT-SCHEMA-CRUD

func resourceJDCloudDiskAttachment() *schema.Resource {

	return &schema.Resource{
		Create: resourceJDCloudDiskAttachmentCreate,
		Read:   resourceJDCloudDiskAttachmentRead,
		Update: resourceJDCloudDiskAttachmentUpdate,
		Delete: resourceJDCloudDiskAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringNoEmpty,
			},
			"disk_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringNoEmpty,
			},
			"auto_delete": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"device_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// This field will be used only in deleting
			// Thus not updated in `Create` and `Read`
			"force_detach": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudDiskAttachmentCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)

	req := apis.NewAttachDiskRequest(config.Region, instanceID, diskID)
	if _, ok := d.GetOk("device_name"); ok {
		req.DeviceName = GetStringAddr(d, "device_name")
	}
	if autoDeleteInterface, ok := d.GetOk("auto_delete"); ok {
		autoDelete := autoDeleteInterface.(bool)
		req.AutoDelete = &autoDelete
	}

	id, e := performDiskAttach(d, meta, req)
	if e != nil {
		return e
	}

	e = diskAttachmentWaiter(d, meta, instanceID, diskID, []string{DISK_ATTACHING}, []string{DISK_ATTACHED})
	if e != nil {
		return e
	}

	d.SetId(id)
	return resourceJDCloudDiskAttachmentRead(d, meta)
}

func resourceJDCloudDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {

	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)

	f := diskAttachmentStatusRefreshFunc(d, meta, instanceID, diskID)
	disk, status, err := f()

	if err != nil {
		return err
	}

	if status != DISK_ATTACHED {
		d.SetId("")
		return nil
	}

	d.Set("instance_id", instanceID)
	d.Set("disk_id", diskID)
	d.Set("device_name", disk.(vm.InstanceDiskAttachment).DeviceName)
	d.Set("auto_delete", disk.(vm.InstanceDiskAttachment).AutoDelete)

	return nil
}

func resourceJDCloudDiskAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("auto_delete") {

		config := meta.(*JDCloudConfig)
		regionID := config.Region
		diskID := GetStringAddr(d, "disk_id")
		autoDelete := d.Get("auto_delete").(bool)
		instanceID := d.Get("instance_id").(string)

		diskAttributeArray := []vm.InstanceDiskAttribute{{DiskId: diskID, AutoDelete: &autoDelete}}
		req := apis.NewModifyInstanceDiskAttributeRequestWithAllParams(regionID, instanceID, diskAttributeArray)
		vmClient := client.NewVmClient(config.Credential)
		resp, err := vmClient.ModifyInstanceDiskAttribute(req)

		if err != nil {
			return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskAttachmentCreate failed %s ", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskAttachmentUpdate,Error code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

	}

	return nil
}

func resourceJDCloudDiskAttachmentDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)
	req := apis.NewDetachDiskRequest(config.Region, instanceID, diskID)

	e := performDiskDetach(d, meta, req)
	if e != nil {
		return e
	}

	e = diskAttachmentWaiter(d, meta, instanceID, diskID, []string{DISK_ATTACHED, DISK_DETACHING}, []string{DISK_DETACHED})
	if e != nil {
		return e
	}

	d.SetId("")
	return nil
}
