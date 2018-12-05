package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"testing"
)

const TestAccRDSInstanceConfig = `
resource "jdcloud_rds_instance" "rds-test-2"{
  instance_name = "xiaohantesting"
  engine = "MySQL"
  engine_version = "5.7"
  instance_class = "db.mysql.s1.micro"
  instance_storage_gb = "20"
  az = "cn-north-1a"
  vpc_id = "vpc-npvvk4wr5j"
  subnet_id = "subnet-j8jrei2981"
  charge_mode = "postpaid_by_usage"
  charge_unit = "month"
  charge_duration = "1"
}
`

func TestAccJDCloudRDSInstance_basic(t *testing.T) {
	var rdsId string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRDSInstanceDestroy(&rdsId),
		Steps: []resource.TestStep{
			{
				Config: TestAccRDSInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					// ROUTE_TABLE_ID validation
					testAccIfRDSInstanceExists("jdcloud_rds.rds-test-2", &rdsId),
				),
			},
		},
	})
}
func testAccIfRDSInstanceExists(resourceName string, resourceId *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {
		resourceStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("we can not find a resource namely:{%s} in terraform.State", resourceName)
		}
		if resourceStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, resource is created but ID not set")
		}
		idStoredLocally := resourceStoredLocally.Primary.ID
		// STEP-2 : Check if RouteTable resource has been created remotely
		config := testAccProvider.Meta().(*JDCloudConfig)
		req := apis.NewDescribeInstanceAttributesRequest(config.Region, idStoredLocally)
		rdsClient := client.NewRdsClient(config.Credential)
		resp, err := rdsClient.DescribeInstanceAttributes(req)
		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] Test failed ,Code:%d, Status:%s ,Message :%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
		localInfo := resourceStoredLocally.Primary.Attributes
		remoteInfo := resp.Result.DbInstanceAttributes
		if localInfo["instance_name"] != remoteInfo.InstanceName {
			return fmt.Errorf("instance_name")
		}
		if localInfo["instance_class"] != remoteInfo.InstanceClass {
			return fmt.Errorf("instance_class")
		}
		if localInfo["internal_domain_name"] != remoteInfo.InternalDomainName {
			return fmt.Errorf("internal_domain_name")
		}
		if localInfo["public_domain_name"] != remoteInfo.PublicDomainName {
			return fmt.Errorf("public_domain_name")
		}
		if localInfo["instance_port"] != remoteInfo.InstancePort {
			return fmt.Errorf("instance_port")
		}
		*resourceId = idStoredLocally
		return nil
	}
}
func testAccRDSInstanceDestroy(resourceId *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {
		if *resourceId == "" {
			return fmt.Errorf("resource Id appears to be empty")
		}
		config := testAccProvider.Meta().(*JDCloudConfig)
		req := apis.NewDescribeInstanceAttributesRequest(config.Region, *resourceId)
		rdsClient := client.NewRdsClient(config.Credential)
		resp, err := rdsClient.DescribeInstanceAttributes(req)
		if err != nil {
			return err
		}
		if resp.Result.DbInstanceAttributes.InstanceStatus != "" {
			return fmt.Errorf("[ERROR] resource still exists,check position-4")
		}
		return nil
	}
}
