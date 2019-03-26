package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"time"
)

func resourceJDCloudNetworkSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudNetworkSecurityGroupCreate,
		Read:   resourceJDCloudNetworkSecurityGroupRead,
		Update: resourceJDCloudNetworkSecurityGroupUpdate,
		Delete: resourceJDCloudNetworkSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_security_group_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudNetworkSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcId := d.Get("vpc_id").(string)
	networkSecurityGroupName := d.Get("network_security_group_name").(string)

	vpcClient := client.NewVpcClient(config.Credential)
	rq := apis.NewCreateNetworkSecurityGroupRequest(config.Region, vpcId, networkSecurityGroupName)
	if descriptionInterface, ok := d.GetOk("description"); ok {
		description := descriptionInterface.(string)
		rq.Description = &description
	}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := vpcClient.CreateNetworkSecurityGroup(rq)

		if err == nil {
			d.SetId(resp.Result.NetworkSecurityGroupId)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func resourceJDCloudNetworkSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	sgClient := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeNetworkSecurityGroupRequest(config.Region, d.Id())

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := sgClient.DescribeNetworkSecurityGroup(req)

		if err == nil {
			d.Set("description", resp.Result.NetworkSecurityGroup.Description)
			d.Set("network_security_group_name", resp.Result.NetworkSecurityGroup.NetworkSecurityGroupName)
			d.Set("vpc_id", resp.Result.NetworkSecurityGroup.VpcId)
			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND {
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

func resourceJDCloudNetworkSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)
	config := meta.(*JDCloudConfig)
	sgClient := client.NewVpcClient(config.Credential)

	if d.HasChange("network_security_group_name") || d.HasChange("description") {

		req := apis.NewModifyNetworkSecurityGroupRequestWithAllParams(config.Region, d.Id(), GetStringAddr(d, "description"), GetStringAddr(d, "network_security_group_name"))

		return resource.Retry(time.Minute, func() *resource.RetryError {

			resp, err := sgClient.ModifyNetworkSecurityGroup(req)

			if err == nil {
				d.SetPartial("network_security_group_name")
				d.SetPartial("description")
				d.Partial(false)
				return nil
			}

			if connectionError(err) {
				return resource.RetryableError(formatConnectionErrorMessage())
			} else {
				return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
			}
		})
	}
	return nil
}

func resourceJDCloudNetworkSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)
	rq := apis.NewDeleteNetworkSecurityGroupRequest(config.Region, d.Id())
	resp, err := vpcClient.DeleteNetworkSecurityGroup(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.SetId("")
	return nil
}
