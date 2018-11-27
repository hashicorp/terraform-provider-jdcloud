package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"strconv"
	"testing"
)


const TestAccSecurityGroupRuleConfig = `
resource "jdcloud_network_security_group_rules" "sg-rule-TEST-1"{
	network_security_group_id = "sg-r7ych5nw90"
	add_security_group_rules = [{
		address_prefix =  "0.0.0.0/0"
		direction = "0"
		from_port = "8000"
		protocol = "300"
		to_port = "8900"
	}]
}
`

func TestAccJDCloudSecurityGroupRule_basic(t *testing.T){

	// This SecurityGroupRule ID is used to create and verify subnet
	// Currently declared but assigned values later
	var SecurityGroupRuleId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckSecurityGroupRuleDestroy(&SecurityGroupRuleId),
		Steps: []resource.TestStep{
			{
				Config: TestAccSecurityGroupRuleConfig,
				Check: resource.ComposeTestCheckFunc(

					// SecurityGroupRuleId verification
					testAccIfSecurityGroupRuleExists("jdcloud_network_security_group_rules.sg-rule-TEST-1", &SecurityGroupRuleId),

				),
			},
		},
	})

}



func testAccIfSecurityGroupRuleExists(securityGroupRuleName string,securityGroupRuleId *string) resource.TestCheckFunc{

	return func(stateInfo *terraform.State) error {

		//STEP-1 : Check if securityGroup resource has been created locally
		securityGroupRuleInfoStoredLocally,ok := stateInfo.RootModule().Resources[securityGroupRuleName]
		if ok==false{
			return fmt.Errorf("securityGroup namely {%s} has not been created",securityGroupRuleName)
		}
		if securityGroupRuleInfoStoredLocally.Primary.ID==""{
			return fmt.Errorf("operation failed, resources created but ID not set")
		}
		securityGroupIdStoredLocally := securityGroupRuleInfoStoredLocally.Primary.Attributes["network_security_group_id"]

		//STEP-2 : Check if securityGroup resource has been created remotely
		securityGroupRuleConfig := testAccProvider.Meta().(*JDCloudConfig)
		securityGroupRuleClient := client.NewVpcClient(securityGroupRuleConfig.Credential)

		req := apis.NewDescribeNetworkSecurityGroupRequest(securityGroupRuleConfig.Region,securityGroupIdStoredLocally)
		resp, err := securityGroupRuleClient.DescribeNetworkSecurityGroup(req)

		if err!=nil{
			return err
		}

		sgRulesRemote := resp.Result.NetworkSecurityGroup.SecurityGroupRules
		sgRulesLocal  := securityGroupRuleInfoStoredLocally.Primary

		sgLocalLength,_ := strconv.Atoi(sgRulesLocal.Attributes["add_security_group_rules.#"])

		for i := 0;i<sgLocalLength;i++ {
			flag := false
			addressPrefix :=  sgRulesLocal.Attributes["add_security_group_rules." + strconv.Itoa(i) + ".address_prefix"]
			for _,sgRemote := range sgRulesRemote{
				if addressPrefix == sgRemote.AddressPrefix {
						flag = true
				}
			}
			if flag==false {
				return fmt.Errorf("resource local dues not match remote")
			}
		}

		//  Here subnet resources has been validated to be created locally and
		//  Remotely, next we are going to validate the remaining attributes
		*securityGroupRuleId = securityGroupIdStoredLocally
		return nil
	}
}


func testAccCheckSecurityGroupRuleDestroy(securityGroupRuleId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// This function is not implemented since
		// Delete function of sgRules has not been implemented yet
		return nil
	}
}
