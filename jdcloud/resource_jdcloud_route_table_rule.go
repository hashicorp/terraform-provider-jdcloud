package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpcModels "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
)

func resourceJDCloudRouteTableRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudRouteTableRuleCreate,
		Read:   resourceJDCloudRouteTableRuleRead,
		Update: resourceJDCloudRouteTableRuleUpdate,
		Delete: resourceJDCloudRouteTableRuleDelete,

		Schema: map[string]*schema.Schema{

			"route_table_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"address_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_hop_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"next_hop_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceJDCloudRouteTableRuleCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	vpcClient := client.NewVpcClient(config.Credential)

	var routeTableRuleSpecs []vpcModels.AddRouteTableRules

	var routeTableRuleSpec vpcModels.AddRouteTableRules

	routeTableID := d.Get("route_table_id").(string)
	routeTableRuleSpec.AddressPrefix = d.Get("address_prefix").(string)
	routeTableRuleSpec.NextHopId = d.Get("next_hop_id").(string)
	routeTableRuleSpec.NextHopType = d.Get("next_hop_type").(string)

	if priorityInterface, ok := d.GetOk("priority"); ok {
		priority := priorityInterface.(int)
		routeTableRuleSpec.Priority = &priority
	}

	routeTableRuleSpecs = append(routeTableRuleSpecs, routeTableRuleSpec)

	rq := apis.NewAddRouteTableRulesRequest(config.Region, routeTableID, routeTableRuleSpecs)
	resp, err := vpcClient.AddRouteTableRules(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudRouteTableRuleCreate failed %s ", err.Error())
		return err
	} else if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudRouteTableRuleCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.RequestID)
	return nil
}
func resourceJDCloudRouteTableRuleRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}
func resourceJDCloudRouteTableRuleUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}
func resourceJDCloudRouteTableRuleDelete(d *schema.ResourceData, meta interface{}) error {

	//return errors.New("do not support delete route table rule ")
	//TODO route table rules do not have id,so this func is invalid
	return nil
}
