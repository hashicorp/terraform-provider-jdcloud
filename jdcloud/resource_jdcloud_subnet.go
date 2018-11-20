package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
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
				Description: "Id of vpc",
			},

			"cidr_block": {
				Type:        schema.TypeString,
				Required:    true,
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
	// Be aware that [cidr_block of subnet] must be a subset of the [Vpc cidr_block] it attaches to

	req := apis.NewCreateSubnetRequest(regionId, vpcId, subnetName, addressPrefix)
	resp, err := subnetClient.CreateSubnet(req)

	if err != nil {
		log.Printf("[DEBUG] resourceSubnetCreate failed %s ", err.Error())
		return err
	} else if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceSubnetCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.Result.SubnetId)

	return nil
}

func resourceSubnetRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceSubnetRead(d, m)
}

func resourceSubnetDelete(d *schema.ResourceData, m interface{}) error {
	if _, err := deleteSubnet(d, m); err != nil {
		return fmt.Errorf("Cannot delete subet: %s", err)
	}
	d.SetId("")
	return nil
}
