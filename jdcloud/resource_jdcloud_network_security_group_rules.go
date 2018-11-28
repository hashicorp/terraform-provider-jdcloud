package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"strconv"
)

func resourceJDCloudNetworkSecurityGroupRules() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudNetworkSecurityGroupRulesCreate,
		Read:   resourceJDCloudNetworkSecurityGroupRulesRead,
		Update: resourceJDCloudNetworkSecurityGroupRulesUpdate,
		Delete: resourceJDCloudNetworkSecurityGroupRulesDelete,

		Schema: map[string]*schema.Schema{
			"network_security_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"add_security_group_rules": &schema.Schema{
				Type:     schema.TypeList,
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

func resourceJDCloudNetworkSecurityGroupRulesCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	networkSecurityGroupID := d.Get("network_security_group_id").(string)
	vpcClient := client.NewVpcClient(config.Credential)
	var networkSecurityGroupRuleSpecs []vpc.AddSecurityGroupRules

	v := d.Get("add_security_group_rules")
	for _, vv := range v.([]interface{}) {

		ele := vv.(map[string]interface{})

		var addSecurityGroupRules vpc.AddSecurityGroupRules

		addSecurityGroupRules.AddressPrefix = ele["address_prefix"].(string)
		addSecurityGroupRules.Direction = ele["direction"].(int)
		addSecurityGroupRules.Protocol = ele["protocol"].(int)

		if fromPortInterface, ok := ele["from_port"]; ok {
			fromPort := fromPortInterface.(int)
			addSecurityGroupRules.FromPort = &fromPort
		}

		if toPortInterface, ok := ele["to_port"]; ok {
			toPort := toPortInterface.(int)
			addSecurityGroupRules.FromPort = &toPort
		}

		if descriptionInterface, ok := ele["description"]; ok {
			description := descriptionInterface.(string)
			addSecurityGroupRules.Description = &description
		}
		networkSecurityGroupRuleSpecs = append(networkSecurityGroupRuleSpecs, addSecurityGroupRules)
	}

	rq := apis.NewAddNetworkSecurityGroupRulesRequest(config.Region, networkSecurityGroupID, networkSecurityGroupRuleSpecs)
	resp, err := vpcClient.AddNetworkSecurityGroupRules(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesCreate failed %s ", err.Error())
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	// This step is set since rule ID can not be retrieved via "create"
	resourceJDCloudNetworkSecurityGroupRulesRead(d,meta)

	d.SetId(networkSecurityGroupID)
	return nil
}


func resourceJDCloudNetworkSecurityGroupRulesRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	ruleClient := client.NewVpcClient(config.Credential)
	networkSecurityGroupID := d.Get("network_security_group_id").(string)
	req := apis.NewDescribeNetworkSecurityGroupRequest(config.Region,networkSecurityGroupID)
	resp, err := ruleClient.DescribeNetworkSecurityGroup(req)

	if err!=nil{
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesRead failed %s ", err.Error())
	}
	if resp.Error.Code == 404 {
		d.SetId("")
		return nil
	}
	if resp.Error.Code!=0{
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	sgRules := resp.Result.NetworkSecurityGroup.SecurityGroupRules
	sgRuleArray := make([]map[string]interface{},0,len(sgRules))

	for _,rule := range sgRules {

		sgRule := map[string]interface{}{
			"address_prefix" : rule.AddressPrefix,
			"description" 	 : rule.Description,
			"direction"      : rule.Direction,
			"from_port"      : rule.FromPort,
			"protocol"		 : rule.Protocol,
			"to_port"        : rule.ToPort,
			"rule_id"        : rule.RuleId,
		}

		sgRuleArray = append(sgRuleArray,sgRule)
	}
	d.Set("add_security_group_rules",sgRuleArray)
	return nil
}



func resourceJDCloudNetworkSecurityGroupRulesUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("add_security_group_rules"){

		config := meta.(*JDCloudConfig)
		ruleClient := client.NewVpcClient(config.Credential)

		sgRuleLength := d.Get("add_security_group_rules.#").(int)
		modifySecurityGroupRuleSpecs := make([]vpc.ModifySecurityGroupRules,0,sgRuleLength)

		for i:=0;i<sgRuleLength;i++{

			sgRule := vpc.ModifySecurityGroupRules{
				d.Get("rule_id").(string),
				GetIntAddr(d,"protocol"),
				GetIntAddr(d,"from_port"),
				GetIntAddr(d,"to_port"),
				GetStringAddr(d,"address_prefix"),
				GetStringAddr(d,"description"),
			}

			modifySecurityGroupRuleSpecs = append(modifySecurityGroupRuleSpecs,sgRule)
		}

		req := apis.NewModifyNetworkSecurityGroupRulesRequest(config.Region,d.Id(),modifySecurityGroupRuleSpecs)
		resp,err := ruleClient.ModifyNetworkSecurityGroupRules(req)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesUpdate failed %s ", err.Error())
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}

	return nil
}



func resourceJDCloudNetworkSecurityGroupRulesDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	ruleClient := client.NewVpcClient(config.Credential)

	sgRuleLength := d.Get("add_security_group_rules.#").(int)
	sgRuleIdArray := make([]string,0,sgRuleLength)
	for i:=0;i<sgRuleLength;i++{
		index := strconv.Itoa(i)
		ruleId :=  d.Get("add_security_group_rules."+index+".rule_id").(string)
		sgRuleIdArray = append(sgRuleIdArray,ruleId)
	}

	req := apis.NewRemoveNetworkSecurityGroupRulesRequest(config.Region,d.Id(),sgRuleIdArray)
	resp,err := ruleClient.RemoveNetworkSecurityGroupRules(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesDelete failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRulesDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId("")
	return nil
}

func getSgRuleSpecs(d *schema.ResourceData, m interface{}) ([]vpc.SecurityGroupRule,[]string,error){

	config   := m.(*JDCloudConfig)
	sgClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	sgId     := d.Get("network_security_group_id").(string)

	req 	 := apis.NewDescribeNetworkSecurityGroupRequest(regionId,sgId)
	resp,err := sgClient.DescribeNetworkSecurityGroup(req)

	if err!=nil{
		return nil,nil,err
	}

	sgRulesList := resp.Result.NetworkSecurityGroup.SecurityGroupRules
	sgRuleIdList := make([]string,0,len(sgRulesList))
	for _,item := range sgRulesList {
		sgRuleIdList = append(sgRuleIdList,item.RuleId)
	}

	return sgRulesList,sgRuleIdList,nil
}

func getIdList(previous []string,latest []string)([]string){

	IdList :=  make([]string,0,len(latest)-len(previous))
	for _,latestItem := range latest{

		flag := false
		for _,previousItem := range previous{

			if latestItem == previousItem{

				flag = true
			}
		}
		if !flag{
			IdList = append(IdList,latestItem)
		}
	}
	return IdList
}