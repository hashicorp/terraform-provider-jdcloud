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
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{

			"route_table_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceRouteTableCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	routeClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	vpcId := d.Get("vpc_id").(string)
	tableName := d.Get("route_table_name").(string)
	description := d.Get("description").(string)

	req := apis.NewCreateRouteTableRequestWithAllParams(regionId, vpcId, tableName, &description)
	resp, err := routeClient.CreateRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableCreate failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceRouteTableCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.Result.RouteTableId)
	return nil
}

func resourceRouteTableRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	routeClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())
	resp, err := routeClient.DescribeRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableRead failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceRouteTableRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.Set("route_table_name", resp.Result.RouteTable.RouteTableName)
	d.Set("description", resp.Result.RouteTable.Description)
	d.Set("vpc_id", resp.Result.RouteTable.VpcId)

	return nil
}

func resourceRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
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
			return fmt.Errorf("[ERROR] resourceRouteTableUpdate failed %s ", err.Error())
		}

		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] resourceRouteTableUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}

	return nil
}

func resourceRouteTableDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	routeClient := client.NewVpcClient(config.Credential)

	req := apis.NewDeleteRouteTableRequest(config.Region, d.Id())
	resp, err := routeClient.DeleteRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceRouteTableDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId("")
	return nil
}
