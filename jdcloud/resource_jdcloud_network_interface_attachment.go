package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"log"
)

func resourceJDCloudNetworkInterfaceAttach() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudNetworkInterfaceAttachCreate,
		Read:   resourceJDCloudNetworkInterfaceAttachRead,
		Delete: resourceJDCloudNetworkInterfaceAttachDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringNoEmpty,
			},
			"network_interface_id": &schema.Schema{
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
		},
	}
}

func resourceJDCloudNetworkInterfaceAttachCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	networkInterfaceID := d.Get("network_interface_id").(string)

	vmClient := client.NewVmClient(config.Credential)

	rq := apis.NewAttachNetworkInterfaceRequest(config.Region, instanceID, networkInterfaceID)

	if autoDeleteInterface, ok := d.GetOk("auto_delete"); ok {
		autoDelete := autoDeleteInterface.(bool)
		rq.AutoDelete = &autoDelete
	}

	resp, err := vmClient.AttachNetworkInterface(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceAttachCreate failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceAttachCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.RequestID)

	return nil
}
func resourceJDCloudNetworkInterfaceAttachRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceJDCloudNetworkInterfaceAttachDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	networkInterfaceId := d.Get("network_interface_id").(string)

	vmClient := client.NewVmClient(config.Credential)

	rq := apis.NewDetachNetworkInterfaceRequest(config.Region, instanceID, networkInterfaceId)
	resp, err := vmClient.DetachNetworkInterface(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceAttachDelete failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceAttachDelete  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	return nil
}
