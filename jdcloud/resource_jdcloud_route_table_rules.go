package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"strconv"
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
							Default:  100,
						},
						"rule_id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

/* Key Functions */
func resourceRouteTableRulesCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	routeTableRulesClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	routeTableId := d.Get("route_table_id").(string)
	routeTableRuleSpecsInterface := d.Get("route_table_rule_specs")

	routeTableRuleSpecs := interfaceToStructArray(routeTableRuleSpecsInterface)
	req := apis.NewAddRouteTableRulesRequestWithAllParams(regionId, routeTableId, routeTableRuleSpecs)
	resp, err := routeTableRulesClient.AddRouteTableRules(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesCreate failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	//Rule id can only be retrieved via "read"
	resourceRouteTableRulesRead(d,meta)
	d.SetId(routeTableId)
	return nil
}


func resourceRouteTableRulesRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTableRequest(config.Region, d.Get("route_table_id").(string))
	resp, err := vpcClient.DescribeRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesRead failed %s ", err.Error())
	}

	if resp.Error.Code == 404 {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	ruleArrayInStructForm := resp.Result.RouteTable.RouteTableRules
	ruleArrayInMapForm := make([]map[string]interface{}, 0, len(ruleArrayInStructForm))

	for _, rule := range ruleArrayInStructForm {

		aRuleInCorrectFrom := map[string]interface{}{
			"next_hop_type":  rule.NextHopType,
			"next_hop_id":    rule.NextHopId,
			"address_prefix": rule.AddressPrefix,
			"priority":       rule.Priority,
			"rule_id" :       rule.RuleId,
		}

		ruleArrayInMapForm = append(ruleArrayInMapForm, aRuleInCorrectFrom)
	}

	ruleArrayInMapFormWithoutLocal := append(ruleArrayInMapForm[1:])
	d.Set("route_table_rule_specs", ruleArrayInMapFormWithoutLocal)
	return nil
}


func resourceRouteTableRulesUpdate(d *schema.ResourceData, meta interface{}) error {
	//originalResourceData, latestResourceData := d.GetChange("route_table_rule_specs")
	//d.Set("route_table_rule_specs", originalResourceData)
	//resourceRouteTableRulesDelete(d, m)
	//d.Set("route_table_rule_specs", latestResourceData)
	//resourceRouteTableRulesCreate(d, m)

	if d.HasChange("route_table_rule_specs"){

		config := meta.(*JDCloudConfig)
		vpcClient := client.NewVpcClient(config.Credential)
		arrayLength := d.Get("route_table_rule_specs.#").(int)
		modifyRouteTableRuleSpecs := make([]vpc.ModifyRouteTableRules,0,arrayLength)

		for i:=0;i<arrayLength;i++ {

			rule := vpc.ModifyRouteTableRules{
				d.Get("route_table_rule_specs."+strconv.Itoa(i)+".rule_id").(string),
				GetIntAddr(d,"route_table_rule_specs."+strconv.Itoa(i)+".priority"),
				GetStringAddr(d,"route_table_rule_specs."+strconv.Itoa(i)+".next_hop_type"),
				GetStringAddr(d,"route_table_rule_specs."+strconv.Itoa(i)+".next_hop_id"),
				GetStringAddr(d,"route_table_rule_specs."+strconv.Itoa(i)+".address_prefix"),
			}

			modifyRouteTableRuleSpecs = append(modifyRouteTableRuleSpecs,rule)
		}

		req := apis.NewModifyRouteTableRulesRequest(config.Region,d.Id(),modifyRouteTableRuleSpecs)
		resp,err := vpcClient.ModifyRouteTableRules(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceRouteTableRulesUpdate failed %s ", err.Error())
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceRouteTableRulesUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}

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

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesDelete failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId("")
	return nil
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

	req := apis.NewDescribeRouteTableRequest(regionId, routeTableId)
	resp, err := client.DescribeRouteTable(req)

	if err != nil {
		return nil, errors.New("cant query ruleID_array, reasons not sure,check position-3")
	}

	ruleArray := resp.Result.RouteTable.RouteTableRules
	ruleIdArray := make([]string, 0, len(ruleArray))
	for _, rule := range ruleArray {
		ruleIdArray = append(ruleIdArray, rule.RuleId)
	}
	return ruleIdArray, nil
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