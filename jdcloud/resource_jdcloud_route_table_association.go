package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
)

func resourceJDCloudRouteTableAssociation() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteTableAssociationCreate,
		Read:   resourceRouteTableAssociationRead,
		Update: resourceRouteTableAssociationUpdate,
		Delete: resourceRouteTableAssociationDelete,

		Schema: map[string]*schema.Schema {

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	associationClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	subnetIds := typeSetToStringArray(d.Get("subnet_id").(*schema.Set))

	routeTableId := d.Get("route_table_id").(string)
	req := apis.NewAssociateRouteTableRequest(regionId, routeTableId, subnetIds)
	resp, err := associationClient.AssociateRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationCreate failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
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

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		log.Printf("Resource not found, probably have been deleted")
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceRouteTableAssociationRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	if err:= d.Set("subnet_id", resp.Result.RouteTable.SubnetIds);err!=nil{
		return fmt.Errorf("[ERROR] Failed in resourceRouteTableAssociationRead,reasons: %s",err.Error())
	}

	return nil
}

func resourceRouteTableAssociationUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("subnet_id"){

		pInterface,cInterface := d.GetChange("subnet_id")
		p:= pInterface.(*schema.Set)
		c:= cInterface.(*schema.Set)
		i := p.Intersection(c)

		detachList := typeSetToStringArray(p.Difference(i))
		attachList := typeSetToStringArray(c.Difference(i))

		if result:= performSubnetDetach(d,m,detachList) && performSubnetAttach(d,m,attachList);result==false{
			return fmt.Errorf("[ERROR] resourceRouteTableAssociationUpdate failed")
		}

		d.Set("subnet_id",cInterface)
	}

	return nil
}

func resourceRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)

	subnetIds := typeSetToStringArray(d.Get("subnet_id").(*schema.Set))
	routeTableId := d.Get("route_table_id").(string)

	for _, item := range subnetIds {
		req := apis.NewDisassociateRouteTableRequest(config.Region, routeTableId, item)
		resp, err := disassociationClient.DisassociateRouteTable(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceRouteTableAssociationDelete failed %s ", err.Error())
		}

		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] resourceRouteTableAssociationDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}

	d.SetId("")
	return nil
}

func typeSetToStringArray(set *schema.Set) []string{

	ret := []string{}
	for _,item := range set.List() {
		ret = append(ret,item.(string))
	}
	return ret
}

func performSubnetAttach(d *schema.ResourceData, meta interface{},attachList []string) bool {

	performSuccess := true
	config := meta.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)

	req := apis.NewAssociateRouteTableRequest(config.Region, d.Get("route_table_id").(string), attachList)
	resp, err := disassociationClient.AssociateRouteTable(req)
	if err != nil || resp.Error.Code != REQUEST_COMPLETED{
			performSuccess = false
	}

	return len(attachList)==RESOURCE_EMPTY || performSuccess
}

func performSubnetDetach(d *schema.ResourceData, meta interface{},detachList []string) bool {

	performSuccess := true
	config := meta.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)
	routeTableId := d.Get("route_table_id").(string)

	for _, id := range detachList {

		req := apis.NewDisassociateRouteTableRequest(config.Region, routeTableId, id)
		resp, err := disassociationClient.DisassociateRouteTable(req)
		if err != nil || resp.Error.Code != REQUEST_COMPLETED{
			performSuccess = false
		}
	}

	return performSuccess
}