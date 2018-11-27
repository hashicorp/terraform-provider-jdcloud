package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
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
				Required: true,
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

	if v, ok := d.GetOk("add_security_group_rules"); ok {

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
	}

	//构造请求
	rq := apis.NewAddNetworkSecurityGroupRulesRequest(config.Region, networkSecurityGroupID, networkSecurityGroupRuleSpecs)

	//发送请求
	resp, err := vpcClient.AddNetworkSecurityGroupRules(rq)
	if err != nil {

		log.Printf("[DEBUG] CreateNetworkSecurityGroup failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] CreateNetworkSecurityGroup failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(networkSecurityGroupID)

	return nil
}
func resourceJDCloudNetworkSecurityGroupRulesRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}
func resourceJDCloudNetworkSecurityGroupRulesUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}
func resourceJDCloudNetworkSecurityGroupRulesDelete(d *schema.ResourceData, meta interface{}) error {

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