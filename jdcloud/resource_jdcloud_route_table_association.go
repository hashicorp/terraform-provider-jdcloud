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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name you route table",
			},

			"subnet_id": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Vpc name this route table belong to",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "See comments below",

				// This parameter appears only because we were asked not to set all parameters "ForceNew"
				// However, "ForceNew" seems to be the most intuitive and appropriate way here.
				// In order to set them to be "ForceNew". I add this params while we dont really need it.
			},
		},
	}
}

func resourceRouteTableAssociationCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	associationClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	subnetId_interdface := d.Get("subnet_id").([]interface{})
	subnetIds := InterfaceToStringArray(subnetId_interdface)
	routeTableId := d.Get("route_table_id").(string)

	req := apis.NewAssociateRouteTableRequest(regionId, routeTableId, subnetIds)
	resp, err := associationClient.AssociateRouteTable(req)

	if err != nil {
		return nil
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("Sorry we cannot associate this table as expected : %s", resp.Error)
	}
	d.SetId(routeTableId)
	return nil
}

func resourceRouteTableAssociationRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	associationClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())
	resp, err := associationClient.DescribeRouteTable(req)

	if err != nil {
		return err
	}
	d.Set("subnet_id", resp.Result.RouteTable.SubnetIds)
	return nil
}

func resourceRouteTableAssociationUpdate(d *schema.ResourceData, m interface{}) error {
	origin, latest := d.GetChange("subnet_id")

	d.Set("subnet_id", origin)
	err_ptr := resourceRouteTableAssociationDelete(d, m)
	if err_ptr!= nil {
		return err_ptr
	}

	d.Set("subnet_id", latest)
	err_ptr_2 := resourceRouteTableAssociationCreate(d, m)
	if err_ptr_2 != nil {
		return err_ptr_2
	}

	return nil
}

func resourceRouteTableAssociationDelete(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)

	subnetId := d.Get("subnet_id").([]interface{})
	subnetIds := InterfaceToStringArray(subnetId)
	routeTableId := d.Get("route_table_id").(string)

	for _, item := range subnetIds {
		req := apis.NewDisassociateRouteTableRequest(config.Region, routeTableId, item)
		resp, err := disassociationClient.DisassociateRouteTable(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("We cannot disassociate this route table for you : %s", resp.Error)
		}
	}

	d.SetId("")
	return nil
}
