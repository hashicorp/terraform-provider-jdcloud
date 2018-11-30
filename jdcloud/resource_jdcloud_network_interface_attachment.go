package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vpcApis "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"regexp"
	"time"
)

// TODO verify&waiting not complete ++ 404 in reading process

func resourceJDCloudNetworkInterfaceAttach() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudNetworkInterfaceAttachCreate,
		Read: resourceJDCloudNetworkInterfaceAttachRead,
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

/*
	Its been proved that the id of both will be set
	Immediately after the request has been sent
*/
func resourceJDCloudNetworkInterfaceAttachCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	networkInterfaceID := d.Get("network_interface_id").(string)

	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewAttachNetworkInterfaceRequest(config.Region, instanceID, networkInterfaceID)

	if autoDeleteInterface, ok := d.GetOk("auto_delete"); ok {
		autoDelete := autoDeleteInterface.(bool)
		req.AutoDelete = &autoDelete
	}

	resp, err := vmClient.AttachNetworkInterface(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachCreate failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
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

	vpcClient := vpc.NewVpcClient(config.Credential)
	reqQuery := vpcApis.NewDescribeNetworkInterfaceRequest(config.Region, networkInterfaceId)

	for retryCount := 0; retryCount < 3; retryCount++ {

		resp, err := vmClient.DetachNetworkInterface(rq)
		errorMessage := fmt.Sprintf("%s", err)
		previousTaskNotComplete, _ := regexp.MatchString("Conflict", errorMessage)
		previousTaskNotComplete = resp.Error.Code == 400 || previousTaskNotComplete

		resp2, _ := vpcClient.DescribeNetworkInterface(reqQuery)
		instanceIdQueried := resp2.Result.NetworkInterface.InstanceId
		CurrentTaskNotComplete := instanceIdQueried == instanceID

		if CurrentTaskNotComplete || previousTaskNotComplete {
			time.Sleep(5 * time.Second)
			continue
		}

		if err == nil && resp.Error.Code == 0 {
			d.SetId("")
			return nil
		}

		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachDelete failed %s ", err.Error())
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachDelete  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}

	return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachDelete")
}
