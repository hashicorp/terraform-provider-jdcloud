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

			"subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name you route table",
			},

			"route_table_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Vpc name this route table belong to",
			},
		},
	}
}

func resourceRouteTableAssociationCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	associationClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	subnetId := d.Get("subnet_id").(string)
	routeTableId := d.Get("route_table_id").(string)
	subnetIds := []string{subnetId}

	req := apis.NewAssociateRouteTableRequest(regionId, routeTableId, subnetIds)
	resp, err := associationClient.AssociateRouteTable(req)

	if err != nil {
		return nil
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("Sorry we cannot associate this table as expected : %s", resp.Error)
	}
	d.SetId(resp.RequestID)
	return nil
}

/*
	TODO
	We are supposed to inform the operator that the route table he is currently
	using will be disassociated automatically,  if he wish to associate a new one
*/

func resourceRouteTableAssociationRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRouteTableAssociationUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceRouteTableAssociationRead(d, m)
}

func resourceRouteTableAssociationDelete(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)

	subnetId := d.Get("subnet_id").(string)
	routeTableId := d.Get("route_table_id").(string)
	req := apis.NewDisassociateRouteTableRequest(config.Region, routeTableId, subnetId)
	resp, err := disassociationClient.DisassociateRouteTable(req)

	if err != nil {
		return err
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("We cannot disassociate this route table for you : %s", resp.Error)
	}
	d.SetId("")
	return nil
}
