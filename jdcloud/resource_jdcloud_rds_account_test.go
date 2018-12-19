package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"testing"
)

const TestAccRDSAccountConfig = `
resource "jdcloud_rds_account" "rds-test1"{
  instance_id = "mysql-g0afoqpl6y"
  username = "DevOps"
  password = "DevOps2018"
}
`

func TestAccJDCloudRDSAccount_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRDSAccountDestroy("jdcloud_rds_account.rds-test1"),
		Steps: []resource.TestStep{
			{
				Config: TestAccRDSAccountConfig,
				Check: resource.ComposeTestCheckFunc(

					// ROUTE_TABLE_ID validation
					testAccIfRDSAccountExists("jdcloud_rds_account.rds-test1"),
				),
			},
		},
	})
}

func testAccIfRDSAccountExists(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		resourceStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfRDSAccountExists Failed,we can not find a resource namely:{%s} in terraform.State", resourceName)
		}
		if resourceStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfRDSAccountExists Failed,operation failed, resource is created but ID not set")
		}

		instanceId := resourceStoredLocally.Primary.Attributes["instance_id"]
		userName := resourceStoredLocally.Primary.Attributes["username"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		rdsClient := client.NewRdsClient(config.Credential)

		req := apis.NewDescribeAccountsRequest(config.Region, instanceId)
		resp, err := rdsClient.DescribeAccounts(req)
		remoteInfo := resp.Result.Accounts

		if err != nil {
			return err
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		for _, account := range remoteInfo {
			if userName == account.AccountName {
				return nil
			}
		}

		return fmt.Errorf("[ERROR] Test failed , cannot find certain account")
	}
}

func testAccRDSAccountDestroy(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		instanceId := stateInfo.RootModule().Resources[resourceName].Primary.Attributes["instance_id"]
		userName := stateInfo.RootModule().Resources[resourceName].Primary.Attributes["username"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		rdsClient := client.NewRdsClient(config.Credential)

		req := apis.NewDescribeAccountsRequest(config.Region, instanceId)
		resp, err := rdsClient.DescribeAccounts(req)
		remoteInfo := resp.Result.Accounts

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		for _, account := range remoteInfo {
			if userName == account.AccountName {
				return fmt.Errorf("[ERROR] Test failed , resource still exists")
			}
		}

		return nil
	}
}
