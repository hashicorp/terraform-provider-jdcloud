package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
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
	if _,ok := d.GetOk("device_name");ok{
		req.DeviceName =  GetStringAddr(d,"device_name")
	}
	if autoDeleteInterface,ok := d.GetOk("auto_delete");ok{
		autoDelete := autoDeleteInterface.(bool)
		req.AutoDelete = &autoDelete
	}

	vmClient := client.NewVmClient(config.Credential)
	resp, err := vmClient.AttachDisk(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudDiskAttachmentCreate failed %s ", err.Error())
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudDiskAttachmentCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.RequestID)
	return nil
}


func resourceJDCloudDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {

	// If disk has been found detached from instance
	// Remove this resource locally
	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)

	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region,instanceID)
	resp ,err := vmClient.DescribeInstance(req)

	if err != nil {
		return err
	}
	for _,disk := range resp.Result.Instance.DataDisks{
		if diskID == disk.CloudDisk.DiskId{
			d.Set("auto_delete",disk.AutoDelete)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceJDCloudDiskAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("auto_delete") {

		config := meta.(*JDCloudConfig)
		regionID := config.Region
		diskID := GetStringAddr(d,"disk_id")
		autoDelete := d.Get("auto_delete").(bool)
		instanceID := d.Get("instance_id").(string)

		diskAttributeArray := []vm.InstanceDiskAttribute{vm.InstanceDiskAttribute{diskID,&autoDelete}}
		req := apis.NewModifyInstanceDiskAttributeRequestWithAllParams(regionID,instanceID,diskAttributeArray)
		vmClient := client.NewVmClient(config.Credential)
		resp,err := vmClient.ModifyInstanceDiskAttribute(req)

		if err!=nil{
			return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskAttachmentCreate failed %s ", err.Error())
		}
		if resp.Error.Code!=0{
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
	if forceDetachInterface, ok := d.GetOk("force_detach"); ok {
		forceDetach := forceDetachInterface.(bool)
		req.Force = &forceDetach
	}

	vmClient := client.NewVmClient(config.Credential)
	resp, err := vmClient.DetachDisk(req)

	if err != nil {
		return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskAttachmentDelete failed %s ", err.Error())
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskAttachmentDelete,Error code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}
