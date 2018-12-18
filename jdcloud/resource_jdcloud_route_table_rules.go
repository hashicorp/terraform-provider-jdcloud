package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
)

func getMapIntAddr(i int) *int {
	return &i
}

func getMapStrAddr(s string) *string {
	return &s
}

func ruleIdList(set *schema.Set) []string {

	idList := []string{}
	for _, item := range set.List() {
		m := item.(map[string]interface{})
		idList = append(idList, m["rule_id"].(string))
	}

	return idList
}

func typeSetToStructArray(set *schema.Set) []vpc.AddRouteTableRules {

	rules := []vpc.AddRouteTableRules{}

	for _, item := range set.List() {

		m := item.(map[string]interface{})
		rules = append(rules, vpc.AddRouteTableRules{
			NextHopType:   m["next_hop_type"].(string),
			NextHopId:     m["next_hop_id"].(string),
			AddressPrefix: m["address_prefix"].(string),
			Priority:      getMapIntAddr(m["priority"].(int)),
		})
	}
	return rules
}

func ruleMap(ruleStruct []vpc.RouteTableRule) []map[string]interface{} {

	ruleMap := []map[string]interface{}{}
	for _, rule := range ruleStruct {

		ruleMap = append(ruleMap, map[string]interface{}{
			"next_hop_type":  rule.NextHopType,
			"next_hop_id":    rule.NextHopId,
			"address_prefix": rule.AddressPrefix,
			"priority":       rule.Priority,
			"rule_id":        rule.RuleId,
		})
	}
	return ruleMap
}

func performRuleDetach(d *schema.ResourceData, m interface{}, detachList []string) error {

	config := m.(*JDCloudConfig)
	c := client.NewVpcClient(config.Credential)

	req := apis.NewRemoveRouteTableRulesRequest(config.Region, d.Id(), detachList)
	resp, err := c.RemoveRouteTableRules(req)

	if err != nil {
		return err
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] performRuleDetach code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	return nil
}

func performRuleAttach(d *schema.ResourceData, m interface{}, attachList []vpc.AddRouteTableRules) error {

	config := m.(*JDCloudConfig)
	c := client.NewVpcClient(config.Credential)

	req := apis.NewAddRouteTableRulesRequest(config.Region, d.Id(), attachList)
	resp, err := c.AddRouteTableRules(req)

	if err != nil {
		return err
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] performRuleAttach code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	return nil
}

// -----------------------------------------------------

func resourceJDCloudRouteTableRules() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteTableRulesCreate,
		Read:   resourceRouteTableRulesRead,
		Update: resourceRouteTableRulesUpdate,
		Delete: resourceRouteTableRulesDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule_specs": &schema.Schema{
				Type:     schema.TypeSet,
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
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceRouteTableRulesCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	tableId := d.Get("route_table_id").(string)
	routeTableRulesClient := client.NewVpcClient(config.Credential)

	routeTableRuleSpecs := typeSetToStructArray(d.Get("rule_specs").(*schema.Set))
	req := apis.NewAddRouteTableRulesRequest(config.Region, tableId, routeTableRuleSpecs)
	resp, err := routeTableRulesClient.AddRouteTableRules(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesCreate failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(tableId)
	//Rule id can only be retrieved via "read"
	if err := resourceRouteTableRulesRead(d, m); err != nil {
		return err
	}
	return nil
}

func resourceRouteTableRulesRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())
	resp, err := vpcClient.DescribeRouteTable(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesRead failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	ruleMap := ruleMap(resp.Result.RouteTable.RouteTableRules[1:])
	if err := d.Set("rule_specs", ruleMap); err != nil {
		return err
	}
	return nil
}

func resourceRouteTableRulesUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("rule_specs") {

		pInterface, cInterface := d.GetChange("rule_specs")
		p := pInterface.(*schema.Set)
		c := cInterface.(*schema.Set)
		i := p.Intersection(c)

		detachList := ruleIdList(p.Difference(i))
		attachList := typeSetToStructArray(c.Difference(i))

		if err := performRuleDetach(d, m, detachList); err != nil || len(detachList) != 0 {
			return err
		}
		if err := performRuleAttach(d, m, attachList); err != nil || len(attachList) != 0 {
			return err
		}

		d.Set("rule_specs", cInterface)
	}

	return resourceRouteTableRulesRead(d, m)
}

func resourceRouteTableRulesDelete(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	idList := ruleIdList(d.Get("rule_specs").(*schema.Set))
	routeTableRulesClient := client.NewVpcClient(config.Credential)

	req := apis.NewRemoveRouteTableRulesRequest(config.Region, d.Get("route_table_id").(string), idList)
	resp, err := routeTableRulesClient.RemoveRouteTableRules(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceRouteTableRulesDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId("")
	return nil
}
