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
	TestCase : 1.common stuff
               2. [ChargeMode][Fail] Try to create one with "postpaid_by_usage", not quite sure if they are available
				   -> Rds Database doesn't support "postpaid_by_usage", you can only use postpaid_by_duration
*/

const TestAccRDSInstanceConfigTemplate = `
resource "jdcloud_rds_instance" "tftest"{
  instance_name = "tftesting_name"
  engine = "MySQL"
  engine_version = "5.7"
  instance_class = "%s"
  instance_storage_gb = "%s"
  az = "cn-north-1a"
  vpc_id = "%s"
  subnet_id = "%s"
  charge_mode = "postpaid_by_duration"
  charge_unit = "month"
  charge_duration = "1"
}
`

func generateRDSConfig(class, size string) string {
	return fmt.Sprintf(TestAccRDSInstanceConfigTemplate, class, size, packer_vpc, packer_subnet)
}

func TestAccJDCloudRDSInstance_basic(t *testing.T) {
	var rdsId string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRDSInstanceDestroy(&rdsId),
		Steps: []resource.TestStep{
			{
				Config: generateRDSConfig("db.mysql.s1.micro", "20"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfRDSInstanceExists("jdcloud_rds_instance.tftest", &rdsId),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "instance_name", "tftesting_name"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "engine", "MySQL"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "engine_version", "5.7"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "instance_class", "db.mysql.s1.micro"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "instance_storage_gb", "20"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "vpc_id", packer_vpc),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "subnet_id", packer_subnet),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "charge_mode", "postpaid_by_duration"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "charge_unit", "month"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "charge_duration", "1"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_rds_instance.tftest", "internal_domain_name"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_rds_instance.tftest", "instance_port"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_rds_instance.tftest", "connection_mode"),
				),
			},
			{
				Config: generateRDSConfig("db.mysql.s1.medium", "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfRDSInstanceExists("jdcloud_rds_instance.tftest", &rdsId),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "instance_name", "tftesting_name"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "engine", "MySQL"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "engine_version", "5.7"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "instance_class", "db.mysql.s1.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "instance_storage_gb", "100"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "vpc_id", packer_vpc),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "subnet_id", packer_subnet),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "charge_mode", "postpaid_by_duration"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "charge_unit", "month"),
					resource.TestCheckResourceAttr(
						"jdcloud_rds_instance.tftest", "charge_duration", "1"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_rds_instance.tftest", "internal_domain_name"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_rds_instance.tftest", "instance_port"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_rds_instance.tftest", "connection_mode"),
				),
			},
		},
	})
}

func testAccIfRDSInstanceExists(resourceName string, resourceId *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {

		resourceStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfRDSInstanceExists failed ,we can not find a resource namely:{%s} in terraform.State", resourceName)
		}
		if resourceStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfRDSInstanceExists failed ,operation failed, resource is created but ID not set")
		}
		idStoredLocally := resourceStoredLocally.Primary.ID
		*resourceId = resourceStoredLocally.Primary.ID

		config := testAccProvider.Meta().(*JDCloudConfig)
		req := apis.NewDescribeInstanceAttributesRequest(config.Region, idStoredLocally)
		rdsClient := client.NewRdsClient(config.Credential)
		resp, err := rdsClient.DescribeInstanceAttributes(req)

		if err != nil {
			return err
		}

		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		*resourceId = idStoredLocally
		return nil
	}
}
func testAccRDSInstanceDestroy(resourceId *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {
		if *resourceId == "" {
			return fmt.Errorf("[ERROR] testAccRDSInstanceDestroy failed ,oresource Id appears to be empty")
		}
		config := testAccProvider.Meta().(*JDCloudConfig)
		req := apis.NewDescribeInstanceAttributesRequest(config.Region, *resourceId)
		rdsClient := client.NewRdsClient(config.Credential)
		_, err := rdsClient.DescribeInstanceAttributes(req)
		if err != nil {
			return err
		}
		return nil
	}
}
