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

const TestAccRouteTableAssociationConfig = `
resource "jdcloud_route_table_association" "route-table-association-TEST-1"{
	route_table_id = "rtb-jgso5x1ein"
	subnet_id = ["subnet-j8jrei2981"]
}
`
const TestAccRouteTableAssociationConfigUpdate = `
resource "jdcloud_route_table_association" "route-table-association-TEST-1"{
	route_table_id = "rtb-jgso5x1ein"
	subnet_id = ["subnet-j8jrei2981","subnet-7g3j4bzlnf"]
}
`

func TestAccJDCloudRouteTableAssociation_basic(t *testing.T) {

	var routeTableId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRouteTableAssociationDestroy(&routeTableId),
		Steps: []resource.TestStep{
			{
				Config: TestAccRouteTableAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccIfRouteTableAssociationExists("jdcloud_route_table_association.route-table-association-TEST-1", &routeTableId),
				),
			},
			{
				Config: TestAccRouteTableAssociationConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccIfRouteTableAssociationExists("jdcloud_route_table_association.route-table-association-TEST-1", &routeTableId),
				),
			},
		},
	})
}

func testAccIfRouteTableAssociationExists(name string, routeTableId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		info, ok := stateInfo.RootModule().Resources[name]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfRouteTableAssociationExists Failed,we can not find a RouteTableAssociation namely:{%s} in terraform.State", name)
		}
		if info.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfRouteTableAssociationExists Failed,operation failed, RouteTableAssociation is created but ID not set")
		}
		*routeTableId = info.Primary.ID

		config := testAccProvider.Meta().(*JDCloudConfig)
		c := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeRouteTableRequest(config.Region, *routeTableId)
		resp, err := c.DescribeRouteTable(req)

		if err != nil || resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfRouteTableAssociationExists Failed in reading conf, reasons err: %s, resp:%#v", err.Error(), resp.Error)
		}

		l, _ := strconv.Atoi(info.Primary.Attributes["subnet_id.#"])
		if l != len(resp.Result.RouteTable.SubnetIds) {
			return fmt.Errorf("[ERROR] testAccIfRouteTableAssociationExists Failed,expect to have %d subnet ids, actually getv %d",
				l, len(resp.Result.RouteTable.SubnetIds))
		}

		return nil
	}
}

func testAccRouteTableAssociationDestroy(routeTableId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		//  routeTableId is not supposed to be empty
		if *routeTableId == "" {
			return fmt.Errorf("[ERROR] testAccRouteTableAssociationDestroy Failed,route Table Id appears to be empty")
		}

		routeTableConfig := testAccProvider.Meta().(*JDCloudConfig)
		routeTableClient := client.NewVpcClient(routeTableConfig.Credential)

		routeTableRegion := routeTableConfig.Region
		requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableRegion, *routeTableId)
		responseOnRouteTable, err := routeTableClient.DescribeRouteTable(requestOnRouteTable)

		if err != nil {
			return err
		}
		if len(responseOnRouteTable.Result.RouteTable.SubnetIds) != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccRouteTableAssociationDestroy Failed,routeTableAssociation resource still exists,check position-4")
		}

		return nil
	}
}
