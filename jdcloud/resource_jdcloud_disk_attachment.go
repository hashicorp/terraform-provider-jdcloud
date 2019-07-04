package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	"log"
	"strings"
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

/*
	This function, performDiskAttach = client.AttachDisk(req)
				   performDiskDetach = client.DetachDisk(req)

	Q: Why is it takes 50+ lines just to send a request ????
	A: Error happens:
      *Send_request -> 1. connection Error (bad network)
                       2. Attach/Detach multiple disks to the same instance concurrently -> 400 Conflict
						   				         										 -> 500 Server Error
																				         -> 400 Disk Already Attached (what...)

*/
func performDiskAttach(meta interface{}, diskID, instanceID, deviceName string, autoDelete bool) (requestId string, e error) {

	stateConf := &resource.StateChangeConf{
		Pending: []string{"connection_error", "task_conflict"},
		Target:  []string{"send_request_complete", "disk_already_attached"},
		Refresh: func() (diskItem interface{}, diskState string, e error) {

			config := meta.(*JDCloudConfig)
			c := client.NewVmClient(config.Credential)

			req := apis.NewAttachDiskRequest(config.Region, instanceID, diskID)
			if len(deviceName) > 0 {
				req.DeviceName = &deviceName
			}
			if autoDelete {
				req.AutoDelete = &autoDelete
			}
			resp, err := c.AttachDisk(req)

			if connectionError(err) {
				return "send_request_failed", "connection_error", nil
			}
			if err != nil {
				return nil, "", err
			}
			if resp.Error.Code == REQUEST_COMPLETED {
				requestId = resp.RequestID
				return "send_request_success", "send_request_complete", nil
			}

			// -----------------------------------------------  Concurrent attachment error

			log.Printf("[D] Disk Attachemt error happens, error=%v ,resp=%v", err, resp)
			if resp.Error.Code == REQUEST_INVALID && strings.Contains(resp.Error.Message, DISK_CONCURRENT_ATTACHMENT_ERROR) {
				return "send_request_failed", "task_conflict", nil
			}
			if resp.Error.Code == REQUEST_SERVER_ERROR && strings.Contains(resp.Error.Message, DISK_CONCURRENT_ATTACHMENT_ERROR_2) {
				return "send_request_failed", "task_conflict", nil
			}
			if resp.Error.Code == REQUEST_INVALID && strings.Contains(resp.Error.Message, DISK_ALREADY_ATTACHED) && strings.Contains(resp.Error.Message, diskID) {
				return "send_request_success", "disk_already_attached", nil
			}

			return "send_request_failed",
				"unknown_error",
				fmt.Errorf("Failed in sending disk attachment request, error=%v ,resp=%v", err, resp)
		},
		Delay:      3 * time.Second,
		Timeout:    2 * time.Minute,
		MinTimeout: 1 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return "", fmt.Errorf("[E] Failed in AttachingDisk/WaitingDiskAttaching ,err message:%v", err)
	}
	return requestId, e
}

// This function will send a request of detachment,do not wait,just send, level 0
func performDiskDetach(meta interface{}, diskID, instanceID string, forceDetach bool) error {

	stateConf := &resource.StateChangeConf{
		Pending: []string{"connection_error", "task_conflict"},
		Target:  []string{"send_request_complete", "disk_already_attached"},
		Refresh: func() (diskItem interface{}, diskState string, e error) {

			config := meta.(*JDCloudConfig)
			c := client.NewVmClient(config.Credential)
			req := apis.NewDetachDiskRequest(config.Region, instanceID, diskID)
			if forceDetach {
				req.Force = &forceDetach
			}
			resp, err := c.DetachDisk(req)

			if connectionError(err) {
				return "send_request_failed", "connection_error", nil
			}
			if err != nil {
				return nil, "", err
			}
			if resp.Error.Code == REQUEST_COMPLETED {
				return "send_request_success", "send_request_complete", nil
			}

			// -----------------------------------------------  Concurrent attachment error

			log.Printf("[D] Disk Attachemt error happens, error=%v ,resp=%v", err, resp)
			if resp.Error.Code == REQUEST_INVALID && strings.Contains(resp.Error.Message, DISK_CONCURRENT_ATTACHMENT_ERROR) {
				return "send_request_failed", "task_conflict", nil
			}
			if resp.Error.Code == REQUEST_SERVER_ERROR && strings.Contains(resp.Error.Message, DISK_CONCURRENT_ATTACHMENT_ERROR_2) {
				return "send_request_failed", "task_conflict", nil
			}
			if resp.Error.Code == REQUEST_INVALID && strings.Contains(resp.Error.Message, DISK_ALREADY_ATTACHED) && strings.Contains(resp.Error.Message, diskID) {
				return "send_request_success", "disk_already_attached", nil
			}

			return "send_request_failed",
				"unknown_error",
				fmt.Errorf("Failed in sending disk attachment request, error=%v ,resp=%v", err, resp)
		},
		Delay:      3 * time.Second,
		Timeout:    2 * time.Minute,
		MinTimeout: 1 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in AttachingDisk/WaitingDiskAttaching ,err message:%v", err)
	}
	return nil
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

	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)
	deviceName := ""
	autoDelete := false
	if _, ok := d.GetOk("device_name"); ok {
		deviceName = d.Get("device_name").(string)
	}
	if _, ok := d.GetOk("auto_delete"); ok {
		autoDelete = d.Get("auto_delete").(bool)
	}

	id, e := performDiskAttach(meta, diskID, instanceID, deviceName, autoDelete)
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

	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)
	force_detach := false
	if _, ok := d.GetOk("force_detach"); ok {
		force_detach = d.Get("force_detach").(bool)
	}
	e := performDiskDetach(meta, diskID, instanceID, force_detach)
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
