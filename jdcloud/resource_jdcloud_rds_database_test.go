package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"testing"
)

/*
	TestCase : 1-[Pass].common stuff only. Not yet found any tricky point requires extra attention
*/

const TestAccRDSDatabaseConfig = `
resource "jdcloud_rds_database" "db-TEST"{
  instance_id = "%s"
  db_name = "devops2018"
  character_set = "utf8"
}
`

func generateRDSDatabase() string {
	return fmt.Sprintf(TestAccRDSDatabaseConfig, packer_rds)
}

func TestAccJDCloudRDSDatabase_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRDSDatabaseDestroy("jdcloud_rds_database.db-TEST"),
		Steps: []resource.TestStep{
			{
				Config: generateRDSDatabase(),
				Check: resource.ComposeTestCheckFunc(

					testAccIfRDSDatabaseExists("jdcloud_rds_database.db-TEST"),
					resource.TestCheckResourceAttr("jdcloud_rds_database.db-TEST", "instance_id", packer_rds),
					resource.TestCheckResourceAttr("jdcloud_rds_database.db-TEST", "db_name", "devops2018"),
					resource.TestCheckResourceAttr("jdcloud_rds_database.db-TEST", "character_set", "utf8"),
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

		req := apis.NewDescribeDatabasesRequest(config.Region, instanceId)
		req.SetDbName(dbName)
		resp, err := rdsClient.DescribeDatabases(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		return nil
	}
}

func testAccRDSDatabaseDestroy(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		instanceId := stateInfo.RootModule().Resources[resourceName].Primary.Attributes["instance_id"]
		dbName := stateInfo.RootModule().Resources[resourceName].Primary.Attributes["db_name"]

		config := testAccProvider.Meta()
		resp, err := keepReading(instanceId, config)

		if err != nil {
			return fmt.Errorf("[ERROR] Test failed,error:%#v, Code:%d, Status:%s ,Message :%s", err.Error(), resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		for _, db := range resp.Result.Databases {
			if db.DbName == dbName {
				return fmt.Errorf("[ERROR] Test failed, resource still exists, details: %#v", db)
			}
		}

		return nil
	}
}
