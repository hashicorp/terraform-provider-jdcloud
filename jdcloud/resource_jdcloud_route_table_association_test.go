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
func TestAccJDCloudRouteTableAssociation_basic(t *testing.T) {

	// routeTableId is declared but not assigned any values here
	// It will be assigned value in "testAccIfRouteTableAssociationExists"
	var routeTableId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRouteTableAssociationDestroy(&routeTableId),
		Steps: []resource.TestStep{
			{
				Config: TestAccRouteTableAssociationConfig,
				Check: resource.ComposeTestCheckFunc(

					// Association relationship validation
					testAccIfRouteTableAssociationExists("jdcloud_route_table_association.route-table-association-TEST-1", &routeTableId),

				),
			},
		},
	})
}

func testAccIfRouteTableAssociationExists(routeTableAssociationName string, routeTableId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// STEP-1: Check if RouteTableAssociation resource has been created locally
		routeTableAssociationInfoStoredLocally, ok := stateInfo.RootModule().Resources[routeTableAssociationName]
		if ok == false {
			return fmt.Errorf("we can not find a RouteTableAssociation namely:{%s} in terraform.State", routeTableAssociationName)
		}
		if routeTableAssociationInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, RouteTableAssociation is created but ID not set")
		}
		routeTableIdStoredLocally := routeTableAssociationInfoStoredLocally.Primary.ID

		// STEP-2 : Check if RouteTableAssociation resource has been created remotely
		routeTableAssociationConfig := testAccProvider.Meta().(*JDCloudConfig)
		routeTableAssociationClient := client.NewVpcClient(routeTableAssociationConfig.Credential)

		requestOnRouteTable := apis.NewDescribeRouteTableRequest(routeTableAssociationConfig.Region, routeTableIdStoredLocally)
		responseOnRouteTable, _ := routeTableAssociationClient.DescribeRouteTable(requestOnRouteTable)

		subnetIdArrayStoredRemotely := responseOnRouteTable.Result.RouteTable.SubnetIds
		subnetIdArrayStoredLocally  := routeTableAssociationInfoStoredLocally.Primary
		subnetIdCountLocally,_ := strconv.Atoi(subnetIdArrayStoredLocally.Attributes["subnet_id.#"])

		// STEP-2-1 : Comapre subnet ID stored locally and remotely
		if subnetIdCountLocally!=len(subnetIdArrayStoredRemotely){
			return fmt.Errorf("expect to have %d subnet ids, actually getv %d",
									 subnetIdCountLocally,len(subnetIdArrayStoredRemotely))
		}

		// Verify that each subnet Id recorded locally meets with that remotely
		for i := 0; i < subnetIdCountLocally; i++  {

			flag := false
			localSubnetId := subnetIdArrayStoredLocally.Attributes[("subnet_id." + strconv.Itoa(i))]

			for _,subnetIdRemote := range subnetIdArrayStoredRemotely {
				if localSubnetId == subnetIdRemote {
					flag = true
				}
			}

			if flag==false{
				return fmt.Errorf("association info stored locally and remotely does not match")
			}

		}

		// RouteTable ID has been validated
		// We are going to validate the remaining attributes - name,vpc_id,description
		*routeTableId = routeTableIdStoredLocally
		return nil
	}
}


func testAccRouteTableAssociationDestroy(routeTableId *string) resource.TestCheckFunc {

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

		// Error.Code is supposed to be 404 since RouteTable was actually deleted
		// Meanwhile turns out to be 0, successfully queried. Indicating delete error
		if err != nil {
			return err
		}
		if len(responseOnRouteTable.Result.RouteTable.SubnetIds)!= 0 {
			return fmt.Errorf("routeTableAssociation resource still exists,check position-4")
		}

		return nil
	}
}