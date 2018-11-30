package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/pkg/errors"
	"testing"
)

const TestAccSecurityGroupConfig = `
resource "jdcloud_network_security_group" "sg-TEST-1"{
	description = "test"
	network_security_group_name = "test"
	vpc_id = "vpc-npvvk4wr5j"
}
`

func TestAccJDCloudSecurityGroup_basic(t *testing.T) {

	// This securityGroupId is used to create and verify securityGroup
	// Currently declared but assigned values later
	var securityGroupId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecurityGroupDestroy(&securityGroupId),
		Steps: []resource.TestStep{
			{
				Config: TestAccSecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(

					// securityGroupId verification
					testAccIfSecurityGroupExists("jdcloud_network_security_group.sg-TEST-1", &securityGroupId),
				),
			},
		},
	})

}

func testAccIfSecurityGroupExists(securityGroupName string, securityGroupId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		//STEP-1 : Check if securityGroup resource has been created locally
		securityGroupInfoStoredLocally, ok := stateInfo.RootModule().Resources[securityGroupName]
		if ok == false {
			return fmt.Errorf("securityGroup namely {%s} has not been created", securityGroupName)
		}
		if securityGroupInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, resources created but ID not set")
		}
		securityGroupIdStoredLocally := securityGroupInfoStoredLocally.Primary.ID

		//STEP-2 : Check if securityGroup resource has been created remotely
		securityGroupConfig := testAccProvider.Meta().(*JDCloudConfig)
		securityGroupClient := client.NewVpcClient(securityGroupConfig.Credential)

		req := apis.NewDescribeNetworkSecurityGroupRequest(securityGroupConfig.Region, securityGroupIdStoredLocally)
		resp, err := securityGroupClient.DescribeNetworkSecurityGroup(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("resources created locally but not remotely")
		}

		//  Here securityGroup resources has been validated to be created locally and
		//  Remotely, next we are going to validate the remaining attributes
		*securityGroupId = securityGroupIdStoredLocally
		return nil
	}
}

func testAccCheckSecurityGroupDestroy(securityGroupIdStoredLocally *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// securityGroup ID is not supposed to be empty during testing stage
		if *securityGroupIdStoredLocally == "" {
			return errors.New("securityGroupId is empty")
		}

		securityGroupConfig := testAccProvider.Meta().(*JDCloudConfig)
		securityGroupClient := client.NewVpcClient(securityGroupConfig.Credential)

		req := apis.NewDescribeNetworkSecurityGroupRequest(securityGroupConfig.Region, *securityGroupIdStoredLocally)
		resp, err := securityGroupClient.DescribeNetworkSecurityGroup(req)

		// ErrorCode is supposed to be 404 since the securityGroup has already been deleted
		// err is supposed to be nil pointer since query process shall finish
		if err != nil {
			return err
		}
		if resp.Error.Code != 404 {
			return fmt.Errorf("something wrong happens or resource still exists")
		}
		return nil
	}
}
