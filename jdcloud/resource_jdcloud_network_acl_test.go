package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

const TestAccAclConfig = `
resource "jdcloud_network_acl" "acl-test" {
  vpc_id = "vpc-npvvk4wr5j",
  name = "devops",
  description = "jdcloud"
}
`

func TestAccJDCloudAcl_basic(t *testing.T) {

	var aclId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAclDestroy(&aclId),
		Steps: []resource.TestStep{
			{
				Config: TestAccAclConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfAclExists("jdcloud_network_acl.acl-test", &aclId),
				),
			},
			{
				ResourceName:      "jdcloud_network_acl.acl-test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccIfAclExists(aclName string, aclId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {
		aclInfoStoredLocally, ok := stateInfo.RootModule().Resources[aclName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfAclExists Failed,we can not find a acl namely:{%s} in terraform.State", aclName)
		}
		if aclInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfAclExists Failed,operation failed, acl is created but ID not set")
		}

		*aclId = aclInfoStoredLocally.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(config.Credential)
		req := apis.NewDescribeNetworkAclRequest(config.Region, *aclId)
		resp, err := vpcClient.DescribeNetworkAcl(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfAclExists Failed,according to the ID stored locally,we cannot find any acl on your cloud")
		}

		return nil
	}
}

func testAccAclDestroy(aclId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *aclId == "" {
			return fmt.Errorf("[ERROR] testAccAclDestroy Failed,aclID is empty")
		}

		config := testAccProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(config.Credential)
		req := apis.NewDescribeNetworkAclRequest(config.Region, *aclId)
		resp, err := vpcClient.DescribeNetworkAcl(req)

		if err != nil {
			return err
		}
		if resp.Error.Code == REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccAclDestroy Failed,resource still exists,check position-4")
		}
		return nil
	}
}
