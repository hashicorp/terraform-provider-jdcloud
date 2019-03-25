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

/*
	TestCase : 1-[Pass].common stuff only. Not yet found any tricky point requires extra attention
*/

const TestAccRouteTableRulesConfig = `
resource "jdcloud_route_table_rules" "rule-TEST-1"{
  route_table_id = "rtb-jgso5x1ein"
  rule_specs = [{
    next_hop_type = "internet"
    next_hop_id   = "internet"
    address_prefix= "10.0.0.0/16"
  }]
}
`
const TestAccRouteTableRulesConfigUpdate = `
resource "jdcloud_route_table_rules" "rule-TEST-1"{
  route_table_id = "rtb-jgso5x1ein"
  rule_specs = [{
    next_hop_type = "internet"
    next_hop_id   = "internet"
    address_prefix= "0.0.0.0/0"
  },{
    next_hop_type = "internet"
    next_hop_id   = "internet"
    address_prefix= "10.0.0.0/16"
    priority      = 120
  }]
}
`

func TestAccJDCloudRouteTableRules_basic(t *testing.T) {

	// routeTableRuleId is the key to create,query process
	// Currently declared but assigned values later
	var routeTableId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableRuleDestroy(&routeTableId),
		Steps: []resource.TestStep{
			{
				Config: TestAccRouteTableRulesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccIfRouteTableRuleExists("jdcloud_route_table_rules.rule-TEST-1", &routeTableId),
					// Common
					resource.TestCheckResourceAttr("jdcloud_route_table_rules.rule-TEST-1", "route_table_id", "rtb-jgso5x1ein"),
					// TypeSet item length
					resource.TestCheckResourceAttr("jdcloud_route_table_rules.rule-TEST-1", "rule_specs.#", "1"),
				),
			},
			{
				Config: TestAccRouteTableRulesConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccIfRouteTableRuleExists("jdcloud_route_table_rules.rule-TEST-1", &routeTableId),
					// Common
					resource.TestCheckResourceAttr("jdcloud_route_table_rules.rule-TEST-1", "route_table_id", "rtb-jgso5x1ein"),
					// TypeSet item length
					resource.TestCheckResourceAttr("jdcloud_route_table_rules.rule-TEST-1", "rule_specs.#", "2"),
				),
			},
		},
	})

}

func testAccIfRouteTableRuleExists(ruleName string, routeTableId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		ruleInfoStoredLocally, ok := stateInfo.RootModule().Resources[ruleName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfRouteTableRuleExists Failed,RouteTableRule namely: %s has not been created", ruleName)
		}
		if ruleInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfRouteTableRuleExists Failed,RouteTableRules namely %s created but ID not set", ruleName)
		}
		*routeTableId = ruleInfoStoredLocally.Primary.ID

		config := testAccProvider.Meta().(*JDCloudConfig)
		c := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeRouteTableRequest(config.Region, *routeTableId)
		resp, err := c.DescribeRouteTable(req)

		if err != nil {
			return fmt.Errorf("[ERROR] testAccIfRouteTableRuleExists Failed in reading,reasons:%s", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfRouteTableRuleExists Failed in reading,reasons:%#v", resp.Error)
		}

		ruleCount, _ := strconv.Atoi(ruleInfoStoredLocally.Primary.Attributes["rule_specs.#"])

		// Rules
		if ruleCount != len(resp.Result.RouteTable.RouteTableRules)-1 {
			return fmt.Errorf("[ERROR] testAccIfRouteTableRuleExists Failed,expect to have %d rules remotely,actually get %d(Default case included)",
				ruleCount, len(resp.Result.RouteTable.RouteTableRules))
		}

		return nil
	}
}

func testAccCheckRouteTableRuleDestroy(routeTableId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		//  routeTableId is not supposed to be empty
		if *routeTableId == "" {
			return fmt.Errorf("[ERROR] testAccCheckRouteTableRuleDestroy Failed,route Table Id appears to be empty")
		}

		routeTableConfig := testAccProvider.Meta().(*JDCloudConfig)
		routeTableClient := client.NewVpcClient(routeTableConfig.Credential)

		routeTableRegion := routeTableConfig.Region
		requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableRegion, *routeTableId)
		responseOnRouteTable, err := routeTableClient.DescribeRouteTable(requestOnRouteTable)

		if err != nil {
			return err
		}
		if len(responseOnRouteTable.Result.RouteTable.RouteTableRules) > 1 {
			return fmt.Errorf("[ERROR] testAccCheckRouteTableRuleDestroy Failed,resource still exists check position-5")
		}

		return nil
	}
}
