package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"time"
)

func resourceJDCloudSubnet() *schema.Resource {

	return &schema.Resource{

		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceSubnetCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	conn := client.NewVpcClient(config.Credential)

	req := apis.NewCreateSubnetRequest(config.Region,
		d.Get("vpc_id").(string),
		d.Get("subnet_name").(string),
		d.Get("cidr_block").(string))

	if _, ok := d.GetOk("description"); ok {
		req.Description = GetStringAddr(d, "description")
	}

	e := resource.Retry(20*time.Second, func() *resource.RetryError {

		resp, err := conn.CreateSubnet(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.SubnetId)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

	if e != nil {
		return e
	}
	return resourceSubnetRead(d, m)
}

func resourceSubnetRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	subnetClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeSubnetRequest(config.Region, d.Id())
	resp, err := subnetClient.DescribeSubnet(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceSubnetRead failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceSubnetRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.Set("subnet_name", resp.Result.Subnet.SubnetName)
	d.Set("description", resp.Result.Subnet.Description)
	d.Set("cidr_block", resp.Result.Subnet.AddressPrefix)
	d.Set("vpc_id", resp.Result.Subnet.VpcId)
	return nil
}

func resourceSubnetUpdate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	subnetClient := client.NewVpcClient(config.Credential)

	if d.HasChange("subnet_name") || d.HasChange("description") {
		req := apis.NewModifySubnetRequestWithAllParams(
			config.Region,
			d.Id(),
			GetStringAddr(d, "subnet_name"),
			GetStringAddr(d, "description"),
		)
		resp, err := subnetClient.ModifySubnet(req)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceSubnetUpdate failed %s ", err.Error())
		}

		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] resourceSubnetUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

	}

	return resourceSubnetRead(d, m)
}

func resourceSubnetDelete(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	subnetClient := client.NewVpcClient(config.Credential)

	req := apis.NewDeleteSubnetRequest(config.Region, d.Id())
	resp, err := subnetClient.DeleteSubnet(req)
	if err != nil {
		return fmt.Errorf("[ERROR] resourceSubnetDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceSubnetDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId("")
	return nil
}
