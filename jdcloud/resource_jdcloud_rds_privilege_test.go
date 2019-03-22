package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"strconv"
	"testing"
)

const TestAccRDSPrivilegeConfig = `
resource "jdcloud_rds_privilege" "pri-test" {
  instance_id = "mysql-155pjskhpy"
  username = "jdcloudDevOps"
  account_privilege = [
    {db_name = "jdcloud2017",privilege = "rw"},
    {db_name = "jdcloud2018",privilege = "rw"},
    {db_name = "jdcloud2019",privilege = "ro"},
  ]
}
`
const TestAccRDSPrivilegeConfigUpdate = `
resource "jdcloud_rds_privilege" "pri-test" {
  instance_id = "mysql-155pjskhpy"
  username = "jdcloudDevOps"
  account_privilege = [
    {db_name = "jdcloud2017",privilege = "rw"},
  ]
}
`

func TestAccJDCloudRDSPrivilege_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRDSPrivilegeDestroy("jdcloud_rds_privilege.pri-test"),
		Steps: []resource.TestStep{
			{
				Config: TestAccRDSPrivilegeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccIfRDSPrivilegeExists("jdcloud_rds_privilege.pri-test"),
					resource.TestCheckResourceAttr("jdcloud_rds_privilege.pri-test", "instance_id", "mysql-155pjskhpy"),
					resource.TestCheckResourceAttr("jdcloud_rds_privilege.pri-test", "username", "jdcloudDevOps"),
					resource.TestCheckResourceAttr("jdcloud_rds_privilege.pri-test", "account_privilege.#", "3"),
				),
			},
		},
	})
}

func testAccIfRDSPrivilegeExists(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		resourceStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfRDSPrivilegeExists Failed,we can not find a resource namely:{%s} in terraform.State", resourceName)
		}
		if resourceStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfRDSPrivilegeExists Failed,operation failed, resource is created but ID not set")
		}

		instanceId := resourceStoredLocally.Primary.Attributes["instance_id"]
		userName := resourceStoredLocally.Primary.Attributes["username"]
		privLength, _ := strconv.Atoi(resourceStoredLocally.Primary.Attributes["account_privilege.#"])

		config := testAccProvider.Meta().(*JDCloudConfig)
		rdsClient := client.NewRdsClient(config.Credential)

		req := apis.NewDescribeAccountsRequest(config.Region, instanceId)
		resp, err := rdsClient.DescribeAccounts(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		for _, acc := range resp.Result.Accounts {
			if userName == acc.AccountName && privLength == len(acc.AccountPrivileges) {
				return nil
			}
		}

		return fmt.Errorf("[ERROR] Test failed ,certain resource not found")
	}
}

func testAccRDSPrivilegeDestroy(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		resourceStoredLocally, _ := stateInfo.RootModule().Resources[resourceName]
		instanceId := resourceStoredLocally.Primary.Attributes["instance_id"]
		userName := resourceStoredLocally.Primary.Attributes["username"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		rdsClient := client.NewRdsClient(config.Credential)

		req := apis.NewDescribeAccountsRequest(config.Region, instanceId)
		resp, err := rdsClient.DescribeAccounts(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		for _, acc := range resp.Result.Accounts {
			if userName == acc.AccountName && len(acc.AccountPrivileges) != 0 {
				return fmt.Errorf("[ERROR] Test failed ,certain resource still exists")
			}
		}

		return nil
	}
}
