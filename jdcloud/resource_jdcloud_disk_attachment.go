package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"log"
)

func resourceJDCloudDiskAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudDiskAttachmentCreate,
		Read:   resourceJDCloudDiskAttachmentRead,
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
				ForceNew: true,
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

	vmClient := client.NewVmClient(config.Credential)

	//构造请求
	rq := apis.NewAttachDiskRequest(config.Region, instanceID, diskID)

	//发送请求
	resp, err := vmClient.AttachDisk(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudDiskAttachmentCreate failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudDiskAttachmentCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.RequestID)

	return nil
}
func resourceJDCloudDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceJDCloudDiskAttachmentDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	diskID := d.Get("disk_id").(string)

	vmClient := client.NewVmClient(config.Credential)

	//构造请求
	rq := apis.NewDetachDiskRequest(config.Region, instanceID, diskID)

	if forceDetachInterface, ok := d.GetOk("force_detach"); ok {
		forceDetach := forceDetachInterface.(bool)
		rq.Force = &forceDetach
	}

	//发送请求
	resp, err := vmClient.DetachDisk(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudDiskAttachmentDelete failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudDiskAttachmentDelete  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	//TODO 查询确认卸载

	return nil
}
