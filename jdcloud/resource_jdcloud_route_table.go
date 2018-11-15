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
				ForceNew:    true,
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

	config := m.(*JDCloudConfig)
	routeClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())
	resp, err := routeClient.DescribeRouteTable(req)

	if resp.Error.Code == 404 && resp.Error.Status == "NOT_FOUND" {
		d.SetId("")
		return nil
	}
	if err != nil {
		fmt.Errorf("Sorry we can not read this route table :%s", err)
	}

	// In case of vpc_id got modified by other users
	// We were required to update vpc_id to detect accordingly
	d.Set("route_table_name", resp.Result.RouteTable.RouteTableName)
	d.Set("description", resp.Result.RouteTable.Description)
	d.Set("vpc_id", resp.Result.RouteTable.VpcId)

	return nil
}

func resourceRouteTableUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)

	config := m.(*JDCloudConfig)
	routeClient := client.NewVpcClient(config.Credential)

	if d.HasChange("route_table_name") || d.HasChange("description") {
		req := apis.NewModifyRouteTableRequestWithAllParams(
			config.Region,
			d.Id(),
			GetStringAddr(d, "route_table_name"),
			GetStringAddr(d, "description"),
		)
		resp, err := routeClient.ModifyRouteTable(req)
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

func resourceRouteTableDelete(d *schema.ResourceData, m interface{}) error {
	if _, err := deleteRoute(d, m); err != nil {
		return fmt.Errorf("Cannot delete this route table: %s", err)
	}
	d.SetId("")
	return nil
}
