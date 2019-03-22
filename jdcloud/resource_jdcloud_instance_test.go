package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"strconv"
	"testing"
)

const TestAccInstanceTemplate = `
resource "jdcloud_instance" "DevOps" {
  az            = "cn-north-1a"
  instance_name = "%s"
  instance_type = "c.n1.large"
  image_id      = "bba85cab-dfdc-4359-9218-7a2de429dd80"
  password      = "%s"
  description   = "%s"

  subnet_id              = "subnet-j8jrei2981"
  network_interface_name = "jdcloud"
  primary_ip             = "10.0.5.0"
  security_group_ids     = ["sg-ym9yp1egi0"]

  elastic_ip_bandwidth_mbps = 10
  elastic_ip_provider       = "bgp"

  system_disk = {
    disk_category = "local"
    auto_delete   = true
    device_name   = "vda"
    disk_size_gb =  40
  }
}
`

func TestAccJDCloudInstance_basic(t *testing.T) {

	instanceId := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDiskInstanceDestroy("jdcloud_instance.DevOps", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: generateInstanceConfig("TerraformName1", "DevOps2018~", "terraform testing create"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfInstanceExists("jdcloud_instance.DevOps", &instanceId),
				),
			},
			{
				Config: generateInstanceConfig("TerraformName2", "DevOps2018!", "terraform testing update"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfInstanceExists("jdcloud_instance.DevOps", &instanceId),
				),
			},
			{
				ResourceName:      "jdcloud_instance.DevOps",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Currently, verification on disks is not available
func testAccIfInstanceExists(resourceName string, instanceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("testAccIfInstanceExists failed, resource %s unavailable", resourceName)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, resource:%s is created but ID not set", resourceName)
		}

		*instanceId = infoStoredLocally.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstanceRequest(config.Region, *instanceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return fmt.Errorf("testAccIfInstanceExists failed position-1")
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("testAccIfInstanceExists failed position-2")
		}

		localMap := infoStoredLocally.Primary.Attributes
		remoteStruct := resp.Result.Instance

		if remoteStruct.Description != localMap["description"] {
			return fmt.Errorf("testAccIfInstanceExists failed on description")
		}
		if remoteStruct.PrimaryNetworkInterface.NetworkInterface.PrimaryIp.PrivateIpAddress != localMap["primary_ip"] {
			return fmt.Errorf("testAccIfInstanceExists failed on primary ip")
		}
		if remoteStruct.ImageId != localMap["image_id"] {
			return fmt.Errorf("testAccIfInstanceExists failed on Image id")
		}
		if remoteStruct.InstanceName != localMap["instance_name"] {
			return fmt.Errorf("testAccIfInstanceExists failed on instance name ")
		}
		if remoteStruct.InstanceType != localMap["instance_type"] {
			return fmt.Errorf("testAccIfInstanceExists failed on instance type")
		}
		if remoteStruct.SubnetId != localMap["subnet_id"] {
			return fmt.Errorf("testAccIfInstanceExists failed subnet id")
		}
		if len(remoteStruct.KeyNames) != RESOURCE_EMPTY && remoteStruct.KeyNames[0] != localMap["key_names"] {
			return fmt.Errorf("testAccIfInstanceExists failed on key names")
		}
		sgLength, _ := strconv.Atoi(localMap["security_group_ids.#"])
		if len(remoteStruct.PrimaryNetworkInterface.NetworkInterface.SecurityGroups) != sgLength {
			return fmt.Errorf("testAccIfInstanceExists failed on security group")
		}

		return nil
	}
}

func testAccDiskInstanceDestroy(resourceName string, instanceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, _ := stateInfo.RootModule().Resources[resourceName]
		*instanceId = infoStoredLocally.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstanceRequest(config.Region, *instanceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return fmt.Errorf("failed in deleting certain resources position-10")
		}

		if resp.Error.Code == REQUEST_COMPLETED {
			return fmt.Errorf("failed in deleting certain resources position-11 ,code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		return nil
	}
}

func generateInstanceConfig(instanceName, password, description string) string {
	return fmt.Sprintf(TestAccInstanceTemplate, instanceName, password, description)
}
