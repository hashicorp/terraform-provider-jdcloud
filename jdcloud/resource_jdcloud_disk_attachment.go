package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
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
	d.Partial(true)
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
	resp, err := vmClient.AttachDisk(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudDiskAttachmentCreate failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudDiskAttachmentCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	if errAttaching := waitForDiskAttaching(d, meta, instanceID, diskID, DISK_ATTACHED); errAttaching != nil {
		return fmt.Errorf("[ERROR] failed in attaching disk,reasons: %s", errAttaching.Error())
	}

	d.SetPartial("disk_id")
	d.SetPartial("instance_id")
	d.SetPartial("device_name")
	d.SetPartial("auto_delete")

	d.Partial(false)
	d.SetId(resp.RequestID)
	return nil
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

	// If the instance has been deleted already, remove attachment info from local state
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

	// If this disk is not found attaching on this Instance
	// It is supposed to be detached already.Remove it from local state.
	d.SetId("")
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
	d.Partial(true)

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)
	req := apis.NewDetachDiskRequest(config.Region, instanceID, diskID)
	if forceDetachInterface, ok := d.GetOk("force_detach"); ok {
		forceDetach := forceDetachInterface.(bool)
		req.Force = &forceDetach
	}

	vmClient := client.NewVmClient(config.Credential)
	resp, err := vmClient.DetachDisk(req)

	if err != nil {
		return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskAttachmentDelete failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskAttachmentDelete,Error code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
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

	currentTime := int(time.Now().Unix())
	config := meta.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, instanceId)
	reconnectCount := 0

	for {

		time.Sleep(3 * time.Second)
		resp, err := vmClient.DescribeInstance(req)

		for _, disk := range resp.Result.Instance.DataDisks {
			if diskId == disk.CloudDisk.DiskId {
				if disk.Status == expectedStatus {
					return nil
				}
				break // Jump out to Position-A
			}
		}
		// Position-A

		if int(time.Now().Unix())-currentTime > DISK_ATTACHMENT_TIMEOUT {
			return fmt.Errorf("[ERROR] resourceJDCloudDiskAttachment failed, timeout")
		}

		if err != nil {
			if reconnectCount > MAX_RECONNECT_COUNT {
				return fmt.Errorf("[ERROR] resourceJDCloudDiskAttachment, MAX_RECONNECT_COUNT Exceeded failed %s ", err.Error())
			}
			reconnectCount++
			continue
		} else {
			reconnectCount = 0
		}

	}
}
