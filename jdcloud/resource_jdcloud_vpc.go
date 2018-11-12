package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

/*
	This function was invoked to build vpc resources
	Currently under testing  --+-- Schema still needs some more information
	    					   +-- Format of CIDR block
*/
func resourceJDCloudVpc() *schema.Resource {

	return &schema.Resource{

		Create: resourceVpcCreate,
		Read:   resourceVpcRead,
		Update: resourceVpcUpdate,
		Delete: resourceVpcDelete,

		Schema: map[string]*schema.Schema{

			"vpc_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name your new vpc",
			},

			"cidr_block": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "String of CIDR block",
			},

			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Enter the descriprion of your vpc",
			},
		},
	}
}

//-------------------------------------------------------------------- Helper Function
func QueryVpcDetail(d *schema.ResourceData, m interface{}) (*apis.DescribeVpcResponse, error) {

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeVpcRequest(config.Region, d.Id())
	resp, err := vpcClient.DescribeVpc(req)

	if resp.Error.Code == 404 && resp.Error.Status == "NOT_FOUND" {
		resp.Result.Vpc.VpcId = ""
	}

	return resp, err
	// This can lead to two different results -> No error but did not found ->    VpcId = ""
	// 											 Query does not even success->    err != nil
	//											 Query succeed 				->    d will be updated
}

func DeleteVpcInstance(d *schema.ResourceData, m interface{}) (*apis.DeleteVpcResponse, error) {
	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewDeleteVpcRequest(config.Region, d.Id())
	resp, err := vpcClient.DeleteVpc(req)

	return resp, err
}

//-------------------------------------------------------------------- Key Function
func resourceVpcCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	regionID := config.Region
	vpcName := d.Get("vpc_name").(string)
	addressPrefix := GetStringAddr(d, "cidr_block")
	description := GetStringAddr(d, "description")

	req := apis.NewCreateVpcRequestWithAllParams(regionID, vpcName, addressPrefix, description)
	resp, err := vpcClient.CreateVpc(req)

	/*	TODO:   addressPrefix and description are indeed optional rather than required,
	select the creation function properly according to configuration file.
	*/

	/*
		response consists of:

				1. RequestID(string)
				2. Result.VpcId(string)
				3. Error.Code(int) / Error.Status(string) / Error.Message(string)

	*/

	if err != nil {
		return err
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("%s", resp.Error)
	}
	d.SetId(resp.Result.VpcId)

	return nil
}

func resourceVpcRead(d *schema.ResourceData, m interface{}) error {
	vpcInstanceDetail, err := QueryVpcDetail(d, m)
	if err != nil {

		// When the vpc has been deleted, this won't lead to error
		// We are going to set the id to 0 under such scenario
		if vpcInstanceDetail.Result.Vpc.VpcId == "" {
			d.SetId("")
			return nil
		}

		// This means the query does not even success
		// We should report an error under such scenario
		return fmt.Errorf("Query vpc fail %s", err)
	}

	d.Set("vpc_name", vpcInstanceDetail.Result.Vpc.VpcName)
	d.Set("cidr_block", vpcInstanceDetail.Result.Vpc.AddressPrefix)
	d.Set("description", vpcInstanceDetail.Result.Vpc.Description)

	return nil
}

func resourceVpcUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceVpcRead(d, m)
}

func resourceVpcDelete(d *schema.ResourceData, m interface{}) error {
	if _, err := DeleteVpcInstance(d, m); err != nil {
		return fmt.Errorf("Delete vpc fail %s", err)
	}
	d.SetId("")
	return nil
}
