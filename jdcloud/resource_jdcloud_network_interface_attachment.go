package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vpcApis "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
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
	req := apis.NewAttachNetworkInterfaceRequest(config.Region, instanceID, networkInterfaceID)

	if autoDeleteInterface, ok := d.GetOk("auto_delete"); ok {
		autoDelete := autoDeleteInterface.(bool)
		req.AutoDelete = &autoDelete
	}

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := vmClient.AttachNetworkInterface(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.RequestID)
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func resourceJDCloudNetworkInterfaceAttachRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	networkInterfaceId := d.Get("network_interface_id").(string)

	vpcClient := vpc.NewVpcClient(config.Credential)
	req := vpcApis.NewDescribeNetworkInterfaceRequest(config.Region, networkInterfaceId)

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := vpcClient.DescribeNetworkInterface(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {

			d.Set("network_interface_id", resp.Result.NetworkInterface.NetworkInterfaceId)
			d.Set("instance_id", resp.Result.NetworkInterface.InstanceId)

			return nil
		}

		if resp.Result.NetworkInterface.InstanceId == "" {
			log.Printf("Resource not found, probably have been deleted")
			d.SetId("")
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

// Both of their ids will be attached immediately after the request has been sent.
func resourceJDCloudNetworkInterfaceAttachDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceId := d.Get("instance_id").(string)
	networkInterfaceId := d.Get("network_interface_id").(string)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDetachNetworkInterfaceRequest(config.Region, instanceId, networkInterfaceId)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DetachNetworkInterface(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}

		if connectionError(err) || resp.Error.Code == REQUEST_INVALID {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

	if err != nil {
		return err
	}

	reqDes := vpcApis.NewDescribeNetworkInterfaceRequest(config.Region, networkInterfaceId)
	vpcClient := vpc.NewVpcClient(config.Credential)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vpcClient.DescribeNetworkInterface(reqDes)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED && resp.Result.NetworkInterface.InstanceId == "" {
			d.SetId("")
			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND {
			log.Printf("Resource not found, probably have been deleted")
			d.SetId("")
			return nil
		}

		if connectionError(err) || resp.Result.NetworkInterface.InstanceId != "" {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

// Discarded - Expected to be removed in the future

func waitForCreatingComplete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceId := d.Get("instance_id").(string)
	networkInterfaceId := d.Get("network_interface_id").(string)

	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDetachNetworkInterfaceRequest(config.Region, instanceId, networkInterfaceId)

	for retryCount := 0; retryCount < MAX_RECONNECT_COUNT; retryCount++ {

		resp, err := vmClient.DetachNetworkInterface(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachDelete Failed, reasons:%s", err.Error())
		}

		if resp.Error.Code == REQUEST_INVALID {
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}
	}

	return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachDelete.Tolerance exceeded in waiting to complete")
}

func waitForDetachComplete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	networkInterfaceId := d.Get("network_interface_id").(string)

	vpcClient := vpc.NewVpcClient(config.Credential)
	req := vpcApis.NewDescribeNetworkInterfaceRequest(config.Region, networkInterfaceId)

	for retryCount := 0; retryCount < MAX_NI_RECONNECT; retryCount++ {

		resp, err := vpcClient.DescribeNetworkInterface(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachDelete Failed, reasons:%s", err.Error())
		}

		if resp.Result.NetworkInterface.InstanceId == "" {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceAttachDelete.Tolerance exceeded in detach")
}
