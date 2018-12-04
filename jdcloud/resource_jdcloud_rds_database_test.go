package jdcloud

import (
"fmt"
"github.com/hashicorp/terraform/helper/resource"
"github.com/hashicorp/terraform/terraform"
"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
"testing"
)

const TestAccRDSDatabaseConfig = `
resource "jdcloud_rds_database" "db-TEST"{
  instance_id = "mysql-g0afoqpl6y"
  db_name = "chenyao666"
  character_set = "utf8"
}
`

func TestAccJDCloudRDSDatabase_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRDSDatabaseDestroy("jdcloud_rds_database.db-TEST"),
		Steps: []resource.TestStep{
			{
				Config: TestAccRDSDatabaseConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfRDSDatabaseExists("jdcloud_rds_database.db-TEST"),
				),
			},
		},
	})
}

func testAccIfRDSDatabaseExists(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		resourceStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("we can not find a resource namely:{%s} in terraform.State", resourceName)
		}
		if resourceStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, resource is created but ID not set")
		}

		instanceId := resourceStoredLocally.Primary.Attributes["instance_id"]
		dbName := resourceStoredLocally.Primary.Attributes["db_name"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		rdsClient := client.NewRdsClient(config.Credential)

		req := apis.NewDescribeDatabasesRequestWithAllParams(config.Region,instanceId,&dbName)
		resp,err := rdsClient.DescribeDatabases(req)

		if err != nil {
			return err
		}
		if resp.Error.Code == 404 {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s",resp.Error.Code,resp.Error.Status,resp.Error.Message)
		}

		return nil
	}
}

func testAccRDSDatabaseDestroy(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		instanceId := stateInfo.RootModule().Resources[resourceName].Primary.Attributes["instance_id"]
		dbName := stateInfo.RootModule().Resources[resourceName].Primary.Attributes["db_name"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		rdsClient := client.NewRdsClient(config.Credential)

		req := apis.NewDescribeDatabasesRequestWithAllParams(config.Region,instanceId,&dbName)
		resp,err := rdsClient.DescribeDatabases(req)

		if err != nil {
			return err
		}
		if resp.Error.Code == 0 {
			return fmt.Errorf("[ERROR] Test failed, resource still exists ,Code:%d, Status:%s ,Message :%s",resp.Error.Code,resp.Error.Status,resp.Error.Message)
		}

		return nil
	}
}
