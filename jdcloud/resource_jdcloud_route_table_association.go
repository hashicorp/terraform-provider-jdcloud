package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

func resourceJDCloudRouteTableAssociation() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteTableAssociationCreate,
		Read:   resourceRouteTableAssociationRead,
		Update: resourceRouteTableAssociationUpdate,
		Delete: resourceRouteTableAssociationDelete,

		Schema: map[string]*schema.Schema{

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	associationClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	subnetIdInterface := d.Get("subnet_id").([]interface{})
	subnetIds := InterfaceToStringArray(subnetIdInterface)
	routeTableId := d.Get("route_table_id").(string)

	req := apis.NewAssociateRouteTableRequest(regionId, routeTableId, subnetIds)
	resp, err := associationClient.AssociateRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationCreate failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(routeTableId)
	return nil
}

func resourceRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	associationClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())
	resp, err := associationClient.DescribeRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationRead failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.Set("subnet_id", resp.Result.RouteTable.SubnetIds)
	return nil
}

func resourceRouteTableAssociationUpdate(d *schema.ResourceData, meta interface{}) error {

	origin, latest := d.GetChange("subnet_id")

	d.Set("subnet_id", origin)
	err := resourceRouteTableAssociationDelete(d, meta)
	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationUpdate failed %s ", err.Error())
	}

	d.Set("subnet_id", latest)
	err = resourceRouteTableAssociationCreate(d, meta)
	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationUpdate failed %s ", err.Error())
	}

	return nil
}

func resourceRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)

	subnetId := d.Get("subnet_id").([]interface{})
	subnetIds := InterfaceToStringArray(subnetId)
	routeTableId := d.Get("route_table_id").(string)

	for _, item := range subnetIds {
		req := apis.NewDisassociateRouteTableRequest(config.Region, routeTableId, item)
		resp, err := disassociationClient.DisassociateRouteTable(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceRouteTableAssociationDelete failed %s ", err.Error())
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceRouteTableAssociationDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}

	d.SetId("")
	return nil
}
