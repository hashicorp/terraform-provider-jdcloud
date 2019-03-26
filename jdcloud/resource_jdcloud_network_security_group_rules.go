package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"time"
)

func typeSetToSgRuleList(s *schema.Set) []vpc.AddSecurityGroupRules {

	sgRules := []vpc.AddSecurityGroupRules{}
	for _, i := range s.List() {

		m := i.(map[string]interface{})
		r := vpc.AddSecurityGroupRules{}
		r.Protocol = m["protocol"].(int)
		r.Direction = m["direction"].(int)
		r.AddressPrefix = m["address_prefix"].(string)
		if m["from_port"] != "" {
			r.FromPort = getMapIntAddr(m["from_port"].(int))
		}
		if m["to_port"] != "" {
			r.ToPort = getMapIntAddr(m["to_port"].(int))
		}
		if m["description"] != "" {
			r.Description = getMapStrAddr(m["description"].(string))
		}

		sgRules = append(sgRules, r)
	}

	return sgRules
}

func performSgRuleAttach(d *schema.ResourceData, m interface{}, s *schema.Set) error {
	d.Partial(true)

	config := m.(*JDCloudConfig)
	conn := client.NewVpcClient(config.Credential)
	req := apis.NewAddNetworkSecurityGroupRulesRequest(config.Region, d.Get("security_group_id").(string), typeSetToSgRuleList(s))

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := conn.AddNetworkSecurityGroupRules(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetPartial("security_group_rules")
			d.Partial(false)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func performSgRuleDetach(d *schema.ResourceData, m interface{}, s *schema.Set) error {

	config := m.(*JDCloudConfig)
	conn := client.NewVpcClient(config.Credential)
	req := apis.NewRemoveNetworkSecurityGroupRulesRequest(config.Region, d.Get("security_group_id").(string), ruleIdList(s))

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := conn.RemoveNetworkSecurityGroupRules(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetPartial("security_group_rules")
			d.Partial(false)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
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
				ForceNew: true,
			},

			"security_group_rules": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
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

	if err := performSgRuleAttach(d, m, d.Get("security_group_rules").(*schema.Set)); err != nil {
		return err
	}

	d.SetId(d.Get("security_group_id").(string))
	
	return resourceJDCloudNetworkSecurityGroupRulesRead(d, m)
}

func resourceJDCloudNetworkSecurityGroupRulesRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	ruleClient := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeNetworkSecurityGroupRequest(config.Region, d.Id())
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

		if err := performSgRuleDetach(d, m, p.Difference(i)); len(detachList) != 0 && err != nil {
			return err
		}
		if err := performSgRuleAttach(d, m, c.Difference(i)); len(attachList) != 0 && err != nil {
			return err
		}

	}

	return resourceJDCloudNetworkSecurityGroupRulesRead(d, m)
}

func resourceJDCloudNetworkSecurityGroupRulesDelete(d *schema.ResourceData, m interface{}) error {

	if err := performSgRuleDetach(d, m, d.Get("security_group_rules").(*schema.Set)); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
