package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	diskApis "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	diskClient "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	"time"
)

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
			},
			"device_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
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

	vmClient := client.NewVmClient(config.Credential)

	err := resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := vmClient.AttachDisk(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

	if err != nil {
		return err
	}

	//Attaching Disk usually takes seconds. Let's wait for it
	reqRefresh := diskApis.NewDescribeDiskRequest(config.Region, diskID)
	c := diskClient.NewDiskClient(config.Credential)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{DISK_AVAILABLE},
		Target:     []string{DISK_ATTACHED},
		Refresh:    diskStatusRefreshFunc(reqRefresh, c),
		Timeout:    3 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in AttachingDisk/Waiting disk,err message:%v", err)
	}

	return resourceJDCloudDiskAttachmentRead(d, meta)
}

func resourceJDCloudDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)

	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, instanceID)
	resp, err := vmClient.DescribeInstance(req)

	if err != nil {
		return err
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudDiskAttachmentRead  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	for _, disk := range resp.Result.Instance.DataDisks {
		if diskID == disk.CloudDisk.DiskId {
			d.Set("auto_delete", disk.AutoDelete)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceJDCloudDiskAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)
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
		d.SetPartial("auto_delete")
	}
	d.Partial(false)
	return nil
}

func resourceJDCloudDiskAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)
	req := apis.NewDetachDiskRequest(config.Region, instanceID, diskID)
	vmClient := client.NewVmClient(config.Credential)

	if forceDetachInterface, ok := d.GetOk("force_detach"); ok {
		forceDetach := forceDetachInterface.(bool)
		req.Force = &forceDetach
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DetachDisk(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	if err != nil {
		return err
	}

	if errDetaching := waitForDiskAttaching(d, meta, instanceID, diskID, DISK_DETACHED); errDetaching != nil {
		return fmt.Errorf("[ERROR] Faield in removing resource, reasons are following :%s", errDetaching.Error())
	}

	d.SetPartial("disk_id")
	d.SetPartial("instance_id")
	d.SetPartial("device_name")
	d.SetPartial("auto_delete")

	d.Partial(false)
	d.SetId("")
	return nil
}

func waitForDiskAttaching(d *schema.ResourceData, meta interface{}, instanceId, diskId string, expectedStatus string) error {

	config := meta.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, instanceId)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DescribeInstance(req)

		found := false

		// Immediately after the disk has been created, even though we've
		// got the disk_id, doesn't mean we can query its detail
		// Probably we have to wait for a few seconds ...
		if resp.Error.Code == REQUEST_COMPLETED {

			for _, disk := range resp.Result.Instance.DataDisks {
				if diskId == disk.CloudDisk.DiskId && disk.Status == expectedStatus {
					found = true
					break
				}
			}
		}

		if err == nil && resp.Error.Code == REQUEST_COMPLETED && found {
			d.Partial(false)
			return nil
		}

		if connectionError(err) || !found {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			d.SetId("")
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}
