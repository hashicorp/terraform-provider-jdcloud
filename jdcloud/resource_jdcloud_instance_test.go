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

const TestAccInstanceConfig = `
resource "jdcloud_instance" "xiaohan" {
  az            = "cn-north-1a"
  instance_name = "xiaohantesting2"
  instance_type = "c.n1.large"
  image_id      = "bba85cab-dfdc-4359-9218-7a2de429dd80"
  password      = "hanwalks1995~"

  subnet_id              = "subnet-j8jrei2981"
  network_interface_name = "xixi"
  primary_ip             = "10.0.2.0"
  security_group_ids     = ["sg-ym9yp1egi0"]
  sanity_check           = 1

  elastic_ip_bandwidth_mbps = 10
  elastic_ip_provider       = "bgp"

  system_disk = {
    disk_category = "local"
    auto_delete   = true
    device_name   = "vda"
    no_device     = true
    disk_size_gb =  200
  }
}
`

func TestAccJDCloudInstance_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDiskInstanceDestroy("jdcloud_instance.xiaohan"),
		Steps: []resource.TestStep{
			{
				Config: TestAccInstanceConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfInstanceExists("jdcloud_instance.xiaohan"),
				),
			},
		},
	})
}

// Currently, verification on disks is not available
func testAccIfInstanceExists(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("testAccIfInstanceExists failed, resource %s unavailable", resourceName)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, resource:%s is created but ID not set", resourceName)
		}

		instanceId := infoStoredLocally.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstanceRequest(config.Region, instanceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return fmt.Errorf("testAccIfInstanceExists failed position-1")
		}
		if resp.Error.Code != 0 {
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
		if len(remoteStruct.KeyNames) != 0 && remoteStruct.KeyNames[0] != localMap["key_names"] {
			return fmt.Errorf("testAccIfInstanceExists failed on key names")
		}
		sgLength, _ := strconv.Atoi(localMap["security_group_ids.#"])
		if len(remoteStruct.PrimaryNetworkInterface.NetworkInterface.SecurityGroups) != sgLength {
			return fmt.Errorf("testAccIfInstanceExists failed on security group")
		}

		return nil
	}
}

func testAccDiskInstanceDestroy(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, _ := stateInfo.RootModule().Resources[resourceName]
		instanceId := infoStoredLocally.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstanceRequest(config.Region, instanceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return fmt.Errorf("failed in deleting certain resources position-10")
		}

		if resp.Error.Code == 0 {
			return fmt.Errorf("failed in deleting certain resources position-11 ,code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		return nil
	}
}
