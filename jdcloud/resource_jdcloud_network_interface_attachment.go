package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"log"
	"regexp"
	"time"
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

	// Restart count is set since deleting an attachment needs some time
	// In case the last one has not been removed we are going to retry
	restart_count := 0
	restart_place:
	resp, err := vmClient.AttachNetworkInterface(rq)
	restart_count ++
	errorMessage := fmt.Sprintf("%s",err)
	previousTaskNotComplete,_ := regexp.MatchString("Conflict",errorMessage)
	previousTaskNotComplete    = resp.Error.Code==400 || previousTaskNotComplete

	if restart_count<2 && previousTaskNotComplete {
		time.Sleep(5*time.Second)
		goto restart_place
	}

	if err != nil {
		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceAttachCreate failed %s ", err.Error())
		return fmt.Errorf("haha")
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

	restart_count := 0
	restart_place:
	restart_count ++
	resp, err := vmClient.DetachNetworkInterface(rq)

	errorMessage := fmt.Sprintf("%s",err)
	previousTaskNotComplete,_ := regexp.MatchString("Conflict",errorMessage)
	previousTaskNotComplete    = resp.Error.Code==400 || previousTaskNotComplete

	if restart_count<2 && previousTaskNotComplete {
		time.Sleep(5*time.Second)
		goto restart_place
	}

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceAttachDelete failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceAttachDelete  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId("")
	return nil
}
