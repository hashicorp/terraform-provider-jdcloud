package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

func resourceJDCloudRouteTable() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteTableCreate,
		Read:   resourceRouteTableRead,
		Update: resourceRouteTableUpdate,
		Delete: resourceRouteTableDelete,

		Schema: map[string]*schema.Schema{

			"route_table_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name you route table",
			},

			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Vpc name this route table belong to",
			},

			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "describe this route table",
			},
		},
	}
}

//---------------------------------------------------- Helper Function
func deleteRoute(d *schema.ResourceData, m interface{}) (*apis.DeleteRouteTableResponse, error) {

	config := m.(*JDCloudConfig)
	routeClient := client.NewVpcClient(config.Credential)

	req := apis.NewDeleteRouteTableRequest(config.Region, d.Id())
	resp, err := routeClient.DeleteRouteTable(req)

	return resp, err
}

//---------------------------------------------------- Key Function
func resourceRouteTableCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	routeClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	vpcId := d.Get("vpc_id").(string)
	table_name := d.Get("route_table_name").(string)
	description := d.Get("description").(string)

	req := apis.NewCreateRouteTableRequestWithAllParams(regionId, vpcId, table_name, &description)
	resp, err := routeClient.CreateRouteTable(req)

	if err != nil {
		return err
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("Can not create route table: %s", resp.Error)
	}

	d.SetId(resp.Result.RouteTableId)
	return nil
}

func resourceRouteTableRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRouteTableUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceRouteTableRead(d, m)
}

func resourceRouteTableDelete(d *schema.ResourceData, m interface{}) error {
	if _, err := deleteRoute(d, m); err != nil {
		return fmt.Errorf("Cannot delete this route table: %s", err)
	}
	d.SetId("")
	return nil
}
