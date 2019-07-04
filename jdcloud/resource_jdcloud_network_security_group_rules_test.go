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

const TestAccSecurityGroupRuleTemplate = `
resource "jdcloud_network_security_group_rules" "sg-TEST-1" {
  security_group_id = "%s"
  security_group_rules {
      address_prefix = "1.2.0.0/16"
      direction = "0"
      from_port = "0"
      protocol = "300"
      to_port = "0"
    }
}
`
const TestAccSecurityGroupRuleTemplateMulti = `
resource "jdcloud_network_security_group_rules" "sg-TEST-1" {
  security_group_id ="%s"
  security_group_rules {
      address_prefix = "1.2.0.0/16"
      direction = "0"
      from_port = "0"
      protocol = "300"
      to_port = "0"
    }
  security_group_rules {
      address_prefix = "4.3.0.0/16"
      direction = "1"
      from_port = "0"
      protocol = "300"
      to_port = "0"
    }
}
`

func generateSGRulesTemplate(template string) string {
	return fmt.Sprintf(template, packer_sg)
}

func TestAccJDCloudSecurityGroupRule_basic(t *testing.T) {

	var SecurityGroupRuleId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecurityGroupRuleDestroy(&SecurityGroupRuleId),
		Steps: []resource.TestStep{
			{
				Config: generateSGRulesTemplate(TestAccSecurityGroupRuleTemplate),
				Check: resource.ComposeTestCheckFunc(
					testAccIfSecurityGroupRuleExists(
						"jdcloud_network_security_group_rules.sg-TEST-1", &SecurityGroupRuleId),
					resource.TestCheckResourceAttr(
						"jdcloud_network_security_group_rules.sg-TEST-1", "security_group_id", packer_sg),

					// Sg rules are arranged in TypeSet form, thus cannot find a way to inspect its detail
					resource.TestCheckResourceAttr(
						"jdcloud_network_security_group_rules.sg-TEST-1", "security_group_rules.#", "1"),
				),
			},
			{
				Config: generateSGRulesTemplate(TestAccSecurityGroupRuleTemplateMulti),
				Check: resource.ComposeTestCheckFunc(
					testAccIfSecurityGroupRuleExists(
						"jdcloud_network_security_group_rules.sg-TEST-1", &SecurityGroupRuleId),
					resource.TestCheckResourceAttr(
						"jdcloud_network_security_group_rules.sg-TEST-1", "security_group_id", packer_sg),
					resource.TestCheckResourceAttr(
						"jdcloud_network_security_group_rules.sg-TEST-1", "security_group_rules.#", "2"),
				),
			},
		},
	})

}

func testAccIfSecurityGroupRuleExists(securityGroupRuleName string, securityGroupRuleId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		securityGroupRuleInfoStoredLocally, ok := stateInfo.RootModule().Resources[securityGroupRuleName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfSecurityGroupRuleExists Failed,securityGroupRule namely {%s} has not been created", securityGroupRuleName)
		}
		if securityGroupRuleInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfSecurityGroupRuleExists Failed,operation failed, resources created but ID not set")
		}
		securityGroupIdStoredLocally := securityGroupRuleInfoStoredLocally.Primary.Attributes["network_security_group_id"]

		securityGroupRuleConfig := testAccProvider.Meta().(*JDCloudConfig)
		securityGroupRuleClient := client.NewVpcClient(securityGroupRuleConfig.Credential)

		req := apis.NewDescribeNetworkSecurityGroupRequest(securityGroupRuleConfig.Region, securityGroupIdStoredLocally)
		resp, err := securityGroupRuleClient.DescribeNetworkSecurityGroup(req)

		if err != nil {
			return err
		}

		sgRulesRemote := resp.Result.NetworkSecurityGroup.SecurityGroupRules
		sgRulesLocal := securityGroupRuleInfoStoredLocally.Primary
		sgLocalLength, _ := strconv.Atoi(sgRulesLocal.Attributes["add_security_group_rules.#"])

		if sgLocalLength != len(sgRulesRemote) {
			return fmt.Errorf("[ERROR] testAccIfSecurityGroupRuleExists Failed,resource local dues not match remote")
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
