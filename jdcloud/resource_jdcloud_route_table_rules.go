package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
)

func resourceJDCloudRouteTableRules() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteTableRulesCreate,
		Read:   resourceRouteTableRulesRead,
		Update: resourceRouteTableRulesUpdate,
		Delete: resourceRouteTableRulesDelete,
		Schema: map[string]*schema.Schema{

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"route_table_rule_specs": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						"next_hop_type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"next_hop_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"address_prefix": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"priority": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

/* Helper Functions*/
func interfaceToStructArray(configInterfaceArray interface{}) []vpc.AddRouteTableRules {
	defaultPriority := 100
	var configArray []vpc.AddRouteTableRules

	for _, configInterface := range configInterfaceArray.([]interface{}) {

		d := configInterface.(map[string]interface{})
		conf := vpc.AddRouteTableRules{
			NextHopType:   d["next_hop_type"].(string),
			NextHopId:     d["next_hop_id"].(string),
			AddressPrefix: d["address_prefix"].(string),
			Priority:      &defaultPriority,
		}
		if priority, ok := d["priority"].(int); ok {
			conf.Priority = &priority
		}
		configArray = append(configArray, conf)
	}
	return configArray
}

func returnRuleIdArray(regionId string, routeTableId string, client *client.VpcClient) ([]string, error) {

	describe_req_RT := apis.NewDescribeRouteTableRequest(regionId, routeTableId)
	resp_RT, err := client.DescribeRouteTable(describe_req_RT)
	if err != nil {
		return nil, errors.New("cant query ruleID_array, reasons not sure,check position-3")
	}

	rule_array := resp_RT.Result.RouteTable.RouteTableRules
	rule_id_array := make([]string, 0, len(rule_array))
	for _, a_single_rule := range rule_array {
		rule_id_array = append(rule_id_array, a_single_rule.RuleId)
	}
	return rule_id_array, nil
}

func equalSliceString(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) == 0 {
		return true
	}

	for _, item := range a {
		flag := false
		for _, item2 := range b {
			if item == item2 {
				flag = true
			}
		}
		if flag == false {
			return false
		}
	}

	for _, item := range b {
		flag := false
		for _, item2 := range a {
			if item == item2 {
				flag = true
			}
		}
		if flag == false {
			return false
		}
	}

	return true
}

func sliceABelongToB(a []string, b []string) bool {
	for _,itemInA := range a{
		flag := false
		for _,itemInB := range b{
			if(itemInA==itemInB){
				flag =true
			}
		}
		if flag==false{
			return false
		}
	}
	return true
}

/* Key Functions */
func resourceRouteTableRulesCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	routeTableRulesClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	routeTableId := d.Get("route_table_id").(string)
	routeTableRuleSpecsInterface := d.Get("route_table_rule_specs")

	routeTableRuleSpecs := interfaceToStructArray(routeTableRuleSpecsInterface)
	req := apis.NewAddRouteTableRulesRequestWithAllParams(regionId, routeTableId, routeTableRuleSpecs)
	resp, err := routeTableRulesClient.AddRouteTableRules(req)

	if err != nil || resp.Error.Code != 0 {
		return fmt.Errorf("failed in adding route table rule,check position-2,%s",resp.Error)
	}

	d.SetId(routeTableId)
	return nil
}

func resourceRouteTableRulesRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	routeTableRulesClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	routeTableId := d.Get("route_table_id").(string)
	describeRequestOnRouteTable := apis.NewDescribeRouteTableRequest(regionId, routeTableId)
	describeResponseOnRouteTable, _ := routeTableRulesClient.DescribeRouteTable(describeRequestOnRouteTable)
	latestRuleListInStructForm := describeResponseOnRouteTable.Result.RouteTable.RouteTableRules

	latestRuleListInCorrectForm := make([]map[string]interface{}, 0, len(latestRuleListInStructForm))
	for _, rule := range latestRuleListInStructForm {

		aRuleInCorrectFrom := map[string]interface{}{
			"next_hop_type":  rule.NextHopType,
			"next_hop_id":    rule.NextHopId,
			"address_prefix": rule.AddressPrefix,
			"priority":       rule.Priority,
		}

		latestRuleListInCorrectForm = append(latestRuleListInCorrectForm, aRuleInCorrectFrom)
	}

	latestRuleListInCorrectFormWithoutLocal := append(latestRuleListInCorrectForm[1:])
	d.Set("route_table_rule_specs", latestRuleListInCorrectFormWithoutLocal)
	return nil
}

func resourceRouteTableRulesUpdate(d *schema.ResourceData, m interface{}) error {
	originalResourceData, latestResourceData := d.GetChange("route_table_rule_specs")
	d.Set("route_table_rule_specs", originalResourceData)
	resourceRouteTableRulesDelete(d, m)
	d.Set("route_table_rule_specs", latestResourceData)
	resourceRouteTableRulesCreate(d, m)
	return nil
}

func resourceRouteTableRulesDelete(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	routeTableRulesClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	routeTableId := d.Get("route_table_id").(string)
	ruleIds, _ := returnRuleIdArray(regionId, routeTableId, routeTableRulesClient)
	ruleIds = append(ruleIds[1:])

	req := apis.NewRemoveRouteTableRulesRequest(regionId, routeTableId, ruleIds)
	resp, err := routeTableRulesClient.RemoveRouteTableRules(req)

	if err != nil || resp.Error.Code != 0 {
		log.Print("%s", resp.Error)
		return errors.New("failed in deleting rules,check position-1")
	}

	d.SetId("")
	return nil
}
