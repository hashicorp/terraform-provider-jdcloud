package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

/*
	TestCase : 1-[Pass].common stuff only. Not yet found any tricky point requires extra attention
*/

const TestAccRouteTableConfigMin = `
resource "jdcloud_route_table" "route-table-TEST-1"{
	route_table_name = "route_table_test"
	vpc_id = "vpc-npvvk4wr5j"
}
`
const TestAccRouteTableConfig = `
resource "jdcloud_route_table" "route-table-TEST-1"{
	route_table_name = "route_table_test"
	vpc_id = "vpc-npvvk4wr5j"
	description = "test"
}
`
const TestAccRouteTableConfigUpdate = `
resource "jdcloud_route_table" "route-table-TEST-1"{
	route_table_name = "route_table_test2"
	vpc_id = "vpc-npvvk4wr5j"
	description = "test with a different name"
}
`

func TestAccJDCloudRouteTable_basic(t *testing.T) {

	var routeTableId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRouteTableDestroy(&routeTableId),
		Steps: []resource.TestStep{
			{
				Config: TestAccRouteTableConfigMin,
				Check: resource.ComposeTestCheckFunc(

					testAccIfRouteTableExists("jdcloud_route_table.route-table-TEST-1", &routeTableId),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "route_table_name", "route_table_test"),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "vpc_id", "vpc-npvvk4wr5j"),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "description", ""),
				),
			},
			{
				Config: TestAccRouteTableConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfRouteTableExists("jdcloud_route_table.route-table-TEST-1", &routeTableId),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "route_table_name", "route_table_test"),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "vpc_id", "vpc-npvvk4wr5j"),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "description", "test"),
				),
			},
			{
				Config: TestAccRouteTableConfigUpdate,
				Check: resource.ComposeTestCheckFunc(

					testAccIfRouteTableExists("jdcloud_route_table.route-table-TEST-1", &routeTableId),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "route_table_name", "route_table_test2"),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "vpc_id", "vpc-npvvk4wr5j"),
					resource.TestCheckResourceAttr("jdcloud_route_table.route-table-TEST-1", "description", "test with a different name"),
				),
			},
			{
				ResourceName:      "jdcloud_route_table.route-table-TEST-1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccIfRouteTableExists(routeTableName string, routeTableId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// STEP-1: Check if RouteTable resource has been created locally
		routeTableInfoStoredLocally, ok := stateInfo.RootModule().Resources[routeTableName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfRouteTableExists Failed,we can not find a RouteTable namely:{%s} in terraform.State", routeTableName)
		}
		if routeTableInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfRouteTableExists Failed,operation failed, RouteTable is created but ID not set")
		}
		routeTableIdStoredLocally := routeTableInfoStoredLocally.Primary.ID

		// STEP-2 : Check if RouteTable resource has been created remotely
		routeTableconfig := testAccProvider.Meta().(*JDCloudConfig)
		routeTableClient := client.NewVpcClient(routeTableconfig.Credential)

		requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableconfig.Region, routeTableIdStoredLocally)
		responseOnRouteTable, err := routeTableClient.DescribeRouteTable(requestOnRouteTable)

		if err != nil {
			return err
		}
		if responseOnRouteTable.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfRouteTableExists Failed,according to the ID stored locally,we cannot find any RouteTable created remotely")
		}

		*routeTableId = routeTableIdStoredLocally
		return nil
	}
}

func testAccRouteTableDestroy(routeTableIdStoredLocally *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *routeTableIdStoredLocally == "" {
			return fmt.Errorf("[ERROR] testAccRouteTableDestroy Failed,route Table Id appears to be empty")
		}

		routeTableConfig := testAccProvider.Meta().(*JDCloudConfig)
		routeTableClient := client.NewVpcClient(routeTableConfig.Credential)

		routeTableRegion := routeTableConfig.Region
		requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableRegion, *routeTableIdStoredLocally)
		responseOnRouteTable, err := routeTableClient.DescribeRouteTable(requestOnRouteTable)

		if err != nil {
			return err
		}
		if responseOnRouteTable.Error.Code == REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccRouteTableDestroy Failed,routeTable resource still exists,check position-4")
		}

		return nil
	}
}
