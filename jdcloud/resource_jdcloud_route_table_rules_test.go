package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

const TestAccRouteTableRulesConfig = `
resource "jdcloud_route_table_rules" "rule-TEST-1"{
	route_table_id = ""
	route_table_rule_specs = [{
		next_hop_type = "internet"
		next_hop_id   = "internet"
		address_prefix= ""
		priority      = 100
	}]
}
`
const ROUTETABLERULESPECS = "route_table_rule_specs"

func TestAccJDCloudRouteTableRules_basic(t *testing.T){

	// routeTableRuleId is the key to create,query process
	// Currently declared but assigned values later
	var routeTableRuleId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckRouteTableRuleDestroy(&routeTableRuleId),
		Steps: []resource.TestStep{
			{
				Config: TestAccRouteTableRulesConfig,
				Check: resource.ComposeTestCheckFunc(

					// SUBNET_ID verification
					testAccIfRouteTableRuleExists("jdcloud_route_table_rules.rule-TEST-1", &routeTableRuleId),

				),
			},
		},
	})

}

func testAccIfRouteTableRuleExists(ruleName string,routeTableRuleId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		//STEP-1 : Check if rule resource has been stored locally
		ruleInfoStoredLocally,ok := stateInfo.RootModule().Resources[ruleName]
		if ok==false {
			return fmt.Errorf("RouteTableRule namely: %s has not been created",ruleName)
		}
		if ruleInfoStoredLocally.Primary.ID==""{
			return fmt.Errorf("RouteTableRules namely %s created but ID not set",ruleName)
		}

		//STEP-2 : Check if rules has been created remotely
		//routeTableConfig := testAccProvider.Meta().(*JDCloudConfig)
		//routeTableClient := client.NewVpcClient(routeTableConfig.Credential)
		//
		//routeTableIdStoredLocally := ruleInfoStoredLocally.Primary.ID
		//routeTableRegion := routeTableConfig.Region
		//requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableRegion, routeTableIdStoredLocally)
		//responseOnRouteTable, _ := routeTableClient.DescribeRouteTable(requestOnRouteTable)
		//
		//ruleListStoredRemotely := responseOnRouteTable.Result.RouteTable.RouteTableRules
		ruleListStoredLocally  := ruleInfoStoredLocally.Primary.Attributes[ROUTETABLERULESPECS]

		return fmt.Errorf("%s",ruleListStoredLocally)

/*		for _,remote := range ruleListStoredRemotely {

			ruleAdress := remote.AddressPrefix

			for _,local := range ruleListStoredLocally {
				if ruleAdress == local["address_prefix"]{

				}
			}


		}*/

	}
}


func testAccCheckRouteTableRuleDestroy(ruleIdStoredLocally *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {
		return nil
	}
}