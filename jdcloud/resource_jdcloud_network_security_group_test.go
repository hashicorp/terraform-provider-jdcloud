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

/*
	TestCase : 1.common stuff only. Not yet found any tricky point requires extra attention
*/
const TestAccSecurityGroupTemplate = `
resource "jdcloud_network_security_group" "TF-TEST"{
	description = "%s"
	network_security_group_name = "%s"
	vpc_id = "vpc-npvvk4wr5j"
}
`

func TestAccJDCloudSecurityGroup_basic(t *testing.T) {

	var securityGroupId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecurityGroupDestroy(&securityGroupId),
		Steps: []resource.TestStep{
			{
				Config: generateSGTemplate("Captain", "JamesMay"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfSecurityGroupExists("jdcloud_network_security_group.TF-TEST", &securityGroupId),
					resource.TestCheckResourceAttr("jdcloud_network_security_group.TF-TEST", "network_security_group_name", "JamesMay"),
					resource.TestCheckResourceAttr("jdcloud_network_security_group.TF-TEST", "description", "Captain"),
				),
			},
			{
				Config: generateSGTemplate("aha", "RichardHammond"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfSecurityGroupExists("jdcloud_network_security_group.TF-TEST", &securityGroupId),
					resource.TestCheckResourceAttr("jdcloud_network_security_group.TF-TEST", "network_security_group_name", "RichardHammond"),
					resource.TestCheckResourceAttr("jdcloud_network_security_group.TF-TEST", "description", "aha"),
				),
			},
			{
				ResourceName:      "jdcloud_network_security_group.TF-TEST",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

}

func testAccIfSecurityGroupExists(securityGroupName string, securityGroupId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		//STEP-1 : Check if securityGroup resource has been created locally
		securityGroupInfoStoredLocally, ok := stateInfo.RootModule().Resources[securityGroupName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfSecurityGroupExists Failed,securityGroup namely {%s} has not been created", securityGroupName)
		}
		if securityGroupInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfSecurityGroupExists Failed,operation failed, resources created but ID not set")
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
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfSecurityGroupExists Failed,resources created locally but not remotely")
		}

		*securityGroupId = securityGroupIdStoredLocally
		return nil
	}
}

func testAccCheckSecurityGroupDestroy(securityGroupIdStoredLocally *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// securityGroup ID is not supposed to be empty during testing stage
		if *securityGroupIdStoredLocally == "" {
			return errors.New("[ERROR] testAccCheckSecurityGroupDestroy Failed,securityGroupId is empty")
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
		if resp.Error.Code != RESOURCE_NOT_FOUND {
			return fmt.Errorf("[ERROR] testAccCheckSecurityGroupDestroy Failed,something wrong happens or resource still exists")
		}
		return nil
	}
}

func generateSGTemplate(des, name string) string {
	return fmt.Sprintf(TestAccSecurityGroupTemplate, des, name)
}
