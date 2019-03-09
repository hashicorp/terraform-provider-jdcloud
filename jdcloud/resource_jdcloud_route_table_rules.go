package jdcloud

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"time"
)

// TODO
// minimum amount
// Validate Func
// Supress Func
// Sensitive Func

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

func typeSetToRouteRuleArray(set *schema.Set) []vpc.AddRouteTableRules {

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

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := c.RemoveRouteTableRules(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func performRuleAttach(d *schema.ResourceData, m interface{}, attachList []vpc.AddRouteTableRules) error {

	config := m.(*JDCloudConfig)
	c := client.NewVpcClient(config.Credential)
	tableId := d.Get("route_table_id").(string)
	req := apis.NewAddRouteTableRulesRequest(config.Region, tableId, attachList)

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := c.AddRouteTableRules(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

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
				MinItems: 1,
			},
		},
	}
}

func resourceRouteTableRulesCreate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)

	tableId := d.Get("route_table_id").(string)
	attachList := typeSetToRouteRuleArray(d.Get("rule_specs").(*schema.Set))

	if err := performRuleAttach(d, m, attachList); err != nil {
		return err
	}

	d.SetPartial("route_table_id")
	d.SetId(tableId)

	//Rule id can only be retrieved via "read"
	if err := resourceRouteTableRulesRead(d, m); err != nil {
		d.SetId("")
		return err
	}

	d.SetPartial("rule_specs")
	d.Partial(false)
	return nil
}

func resourceRouteTableRulesRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := vpcClient.DescribeRouteTable(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			ruleMap := ruleMap(resp.Result.RouteTable.RouteTableRules[1:])
			if err := d.Set("rule_specs", ruleMap); err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND {
			d.SetId("")
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

}

func resourceRouteTableRulesUpdate(d *schema.ResourceData, m interface{}) error {

	d.Partial(true)
	defer d.Partial(false)

	if d.HasChange("rule_specs") {

		pInterface, cInterface := d.GetChange("rule_specs")
		p := pInterface.(*schema.Set)
		c := cInterface.(*schema.Set)
		i := p.Intersection(c)

		detachList := ruleIdList(p.Difference(i))
		attachList := typeSetToRouteRuleArray(c.Difference(i))

		if err := performRuleDetach(d, m, detachList); err != nil && len(detachList) != 0 {
			return err
		}
		d.SetPartial("rule_specs")

		if err := performRuleAttach(d, m, attachList); err != nil && len(attachList) != 0 {
			return err
		}
		d.SetPartial("rule_specs")

	}

	return resourceRouteTableRulesRead(d, m)
}

func resourceRouteTableRulesDelete(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	idList := ruleIdList(d.Get("rule_specs").(*schema.Set))
	routeTableRulesClient := client.NewVpcClient(config.Credential)

	req := apis.NewRemoveRouteTableRulesRequest(config.Region, d.Get("route_table_id").(string), idList)

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := routeTableRulesClient.RemoveRouteTableRules(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId("")
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}
