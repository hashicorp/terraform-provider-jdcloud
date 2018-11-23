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

const TestAccRouteTableRulesConfig = `
resource "jdcloud_route_table_rules" "rule-TEST-1"{
  route_table_id = "rtb-jgso5x1ein"
  route_table_rule_specs = [{
    next_hop_type = "internet"
    next_hop_id   = "internet"
    address_prefix= "10.0.0.0/16"
    priority      = 100
  }]
}
`

func TestAccJDCloudRouteTableRules_basic(t *testing.T){

	// routeTableRuleId is the key to create,query process
	// Currently declared but assigned values later
	var routeTableId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckRouteTableRuleDestroy(&routeTableId),
		Steps: []resource.TestStep{
			{
				Config: TestAccRouteTableRulesConfig,
				Check: resource.ComposeTestCheckFunc(

					// Here we gathered all verification in one function
					// We did this since the info stored remotely cannot be easily get
					testAccIfRouteTableRuleExists("jdcloud_route_table_rules.rule-TEST-1", &routeTableId),

				),
			},
		},
	})

}

func testAccIfRouteTableRuleExists(ruleName string,routeTableId *string) resource.TestCheckFunc {

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

		//STEP-2-1 : Query info on RouteTable first, we did this since info on RouteTableRules
		//can not be queried directly, there is no function namely "describe rules". Hence
		//In order to query info on RouteTableRule we have to query on RouteTable first
		routeTableConfig := testAccProvider.Meta().(*JDCloudConfig)
		routeTableClient := client.NewVpcClient(routeTableConfig.Credential)

		routeTableIdStoredLocally := ruleInfoStoredLocally.Primary.ID
		*routeTableId = routeTableIdStoredLocally
		routeTableRegion := routeTableConfig.Region
		requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableRegion, routeTableIdStoredLocally)
		responseOnRouteTable, _ := routeTableClient.DescribeRouteTable(requestOnRouteTable)

		//STEP-2-2 : Compare stored info on RouteTableRule locally and remotely
		ruleListStoredRemotely := responseOnRouteTable.Result.RouteTable.RouteTableRules
		ruleListStoredRemotelyWithoutDefault := ruleListStoredRemotely[1:]
		ruleCount,_  := strconv.Atoi(ruleInfoStoredLocally.Primary.Attributes["route_table_rule_specs.#"])


		// Compare rule count
		if ruleCount!=len(ruleListStoredRemotelyWithoutDefault){
			return fmt.Errorf("expect to have %d rules remotely,actually get %d(Default case included)",
									 ruleCount+1,len(ruleListStoredRemotelyWithoutDefault))
		}

		// Compare remaining attributes
		attr := ruleInfoStoredLocally.Primary
		for i := 0; i < ruleCount; i++ {
			flag := false
			attrAddress := "route_table_rule_specs." + strconv.Itoa(i) + ".address_prefix"
			attrType   := "route_table_rule_specs." + strconv.Itoa(i) + ".next_hop_type"
			attrId     := "route_table_rule_specs." + strconv.Itoa(i) + ".next_hop_id"
			if ruleListStoredRemotelyWithoutDefault[i].AddressPrefix == attr.Attributes[attrAddress] {
				flag = (ruleListStoredRemotelyWithoutDefault[i].NextHopId == attr.Attributes[attrId]) &&
					   (ruleListStoredRemotelyWithoutDefault[i].NextHopType ==attr.Attributes[attrType] )
			}
			if flag==false{
				return fmt.Errorf("rule info stored locally and remotely does not match")
			}
		}

		//*routeTableId = routeTableIdStoredLocally
		return nil
	}
}


func testAccCheckRouteTableRuleDestroy(routeTableId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		//  routeTableId is not supposed to be empty
		if *routeTableId==""{
			return fmt.Errorf("route Table Id appears to be empty")
		}

		routeTableConfig := testAccProvider.Meta().(*JDCloudConfig)
		routeTableClient := client.NewVpcClient(routeTableConfig.Credential)

		routeTableRegion := routeTableConfig.Region
		requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableRegion, *routeTableId)
		responseOnRouteTable, err := routeTableClient.DescribeRouteTable(requestOnRouteTable)

		if err!=nil{
			return err
		}
		if len(responseOnRouteTable.Result.RouteTable.RouteTableRules) > 1 {
			return fmt.Errorf("resource still exists check position-5")
		}

		return nil
	}
}