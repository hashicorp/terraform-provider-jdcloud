package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
)

func typeSetToSgRuleList(s *schema.Set) []vpc.AddSecurityGroupRules {

	sgRules := []vpc.AddSecurityGroupRules{}
	for _, i := range s.List() {

		m := i.(map[string]interface{})
		r := vpc.AddSecurityGroupRules{}
		r.Protocol = m["protocol"].(int)
		r.Direction = m["direction"].(int)
		r.AddressPrefix = m["address_prefix"].(string)
		if _, ok := m["from_port"]; ok {
			r.FromPort = getMapIntAddr(m["from_port"].(int))
		}
		if _, ok := m["to_port"]; ok {
			r.ToPort = getMapIntAddr(m["to_port"].(int))
		}

		sgRules = append(sgRules, r)
	}

	return sgRules
}

func performSgRuleAttach(d *schema.ResourceData, m interface{}, s *schema.Set) error {

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewAddNetworkSecurityGroupRulesRequest(config.Region, d.Get("security_group_id").(string), typeSetToSgRuleList(s))
	resp, err := vpcClient.AddNetworkSecurityGroupRules(req)

	if err != nil {
		return fmt.Errorf("[ERROR] performSgRuleAttach failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] performSgRuleAttach failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}

func performSgRuleDetach(d *schema.ResourceData, m interface{}, s *schema.Set) error {

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewRemoveNetworkSecurityGroupRulesRequest(config.Region, d.Get("security_group_id").(string), ruleIdList(s))
	resp, err := vpcClient.RemoveNetworkSecurityGroupRules(req)

	if err != nil {
		return fmt.Errorf("[ERROR] performSgRuleDetach failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] performSgRuleDetach failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}

func resourceJDCloudNetworkSecurityGroupRules() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudNetworkSecurityGroupRulesCreate,
		Read:   resourceJDCloudNetworkSecurityGroupRulesRead,
		Update: resourceJDCloudNetworkSecurityGroupRulesUpdate,
		Delete: resourceJDCloudNetworkSecurityGroupRulesDelete,

		Schema: map[string]*schema.Schema{
			"security_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"security_group_rules": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						"address_prefix": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"direction": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"from_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"protocol": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"to_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"rule_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceJDCloudNetworkSecurityGroupRulesCreate(d *schema.ResourceData, m interface{}) error {

	d.Partial(true)

	if err := performSgRuleAttach(d, m, d.Get("security_group_rules").(*schema.Set)); err != nil {
		return err
	}

	d.SetPartial("security_group_rules")
	d.SetPartial("security_group_id")

	if err := resourceJDCloudNetworkSecurityGroupRulesRead(d, m); err != nil {
		return err
	}

	d.Partial(false)
	d.SetId(d.Get("security_group_id").(string))
	return nil
}

func resourceJDCloudNetworkSecurityGroupRulesRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	ruleClient := client.NewVpcClient(config.Credential)
	sgId := d.Get("security_group_id").(string)
	req := apis.NewDescribeNetworkSecurityGroupRequest(config.Region, sgId)
	resp, err := ruleClient.DescribeNetworkSecurityGroup(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesRead failed %s ", err.Error())
	}
	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	sgRules := resp.Result.NetworkSecurityGroup.SecurityGroupRules
	sgRuleArray := make([]map[string]interface{}, 0, len(sgRules))

	for _, rule := range sgRules {

		sgRule := map[string]interface{}{
			"address_prefix": rule.AddressPrefix,
			"description":    rule.Description,
			"direction":      rule.Direction,
			"from_port":      rule.FromPort,
			"protocol":       rule.Protocol,
			"to_port":        rule.ToPort,
			"rule_id":        rule.RuleId,
		}

		sgRuleArray = append(sgRuleArray, sgRule)
	}

	if err := d.Set("security_group_rules", sgRuleArray); err != nil {
		return fmt.Errorf("[ERROR] Failed in resourceJDCloudNetworkSecurityGroupRulesRead,reasons:%s", err.Error())
	}
	return nil
}

func resourceJDCloudNetworkSecurityGroupRulesUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("security_group_rules") {

		pInterface, cInterface := d.GetChange("security_group_rules")
		p := pInterface.(*schema.Set)
		c := cInterface.(*schema.Set)
		i := p.Intersection(c)

		detachList := ruleIdList(p.Difference(i))
		attachList := typeSetToSgRuleList(c.Difference(i))

		if err := performSgRuleDetach(d, m, p.Difference(i)); len(detachList) != 0 || err != nil {
			return err
		}
		if err := performSgRuleAttach(d, m, c.Difference(i)); len(attachList) != 0 || err != nil {
			return err
		}

		d.Set("security_group_rules", cInterface)
	}

	return nil
}

func resourceJDCloudNetworkSecurityGroupRulesDelete(d *schema.ResourceData, m interface{}) error {

	if err := performSgRuleDetach(d, m, d.Get("security_group_rules").(*schema.Set)); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
