package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

func resourceJDCloudSubnet() *schema.Resource {

	return &schema.Resource{

		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,

		Schema: map[string]*schema.Schema{

			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of vpc",
			},

			"cidr_block": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Cidr block,must be the subset of VPC-cidr",
			},

			"subnet_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name your subnet",
			},

			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Describe this subnet",
			},
		},
	}
}

//------------------------------------------------------ Helper Function
func deleteSubnet(d *schema.ResourceData, m interface{}) (*apis.DeleteSubnetResponse, error) {

	config := m.(*JDCloudConfig)
	subnetClient := client.NewVpcClient(config.Credential)

	req := apis.NewDeleteSubnetRequest(config.Region, d.Id())
	resp, err := subnetClient.DeleteSubnet(req)

	return resp, err
}

//------------------------------------------------------ Key Function
func resourceSubnetCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	subnetClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	vpcId := d.Get("vpc_id").(string)
	subnetName := d.Get("subnet_name").(string)
	addressPrefix := d.Get("cidr_block").(string)
	description := GetStringAddr(d, "description")
	// Be aware that [cidr_block of subnet] must be a subset of the [Vpc cidr_block] it attaches to

	req := apis.NewCreateSubnetRequestWithAllParams(regionId, vpcId, subnetName, addressPrefix, nil, description)
	resp, err := subnetClient.CreateSubnet(req)

	if err != nil {
		return err
	}
	if resp.Error.Code != 0 {
		fmt.Errorf("Can not create new subnet: %s", resp.Error)
	}

	d.SetId(resp.Result.SubnetId)

	return nil
}

func resourceSubnetRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	subnetClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeSubnetRequest(config.Region, d.Id())
	resp, err := subnetClient.DescribeSubnet(req)

	if resp.Error.Code == 404 && resp.Error.Status == "NOT_FOUND" {
		d.SetId("")
		return nil
	}
	if err != nil {
		fmt.Errorf("Sorry we can not read this route table :%s", err)
	}

	d.Set("subnet_name", resp.Result.Subnet.SubnetName)
	d.Set("description", resp.Result.Subnet.Description)

	return nil

}

func resourceSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)

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
			return nil
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("We can not make this update: %s", resp.Error)
		}
	}
	d.Partial(false)
	return nil
}

func resourceSubnetDelete(d *schema.ResourceData, m interface{}) error {
	if _, err := deleteSubnet(d, m); err != nil {
		return fmt.Errorf("Cannot delete subet: %s", err)
	}
	d.SetId("")
	return nil
}
