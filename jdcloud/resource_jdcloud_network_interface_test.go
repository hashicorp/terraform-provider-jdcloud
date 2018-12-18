package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
	"strconv"
	"testing"
)

const TestAccNetWorkInterfaceConfig = `
resource "jdcloud_network_interface" "NI-TEST"{
	subnet_id = "subnet-j8jrei2981"
	description = "test"
	az = "cn-north-1"
	network_interface_name = "TerraformTest"
	secondary_ip_addresses = ["10.0.3.0","10.0.4.0"]
	secondary_ip_count = "2"
	security_groups = ["sg-yrd5fa7y55"]
}
`

func TestAccJDCloudNetworkInterface_basic(t *testing.T) {

	var networkInterfaceId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkInterfaceDestroy(&networkInterfaceId),
		Steps: []resource.TestStep{
			{
				Config: TestAccNetWorkInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfNetworkInterfaceExists("jdcloud_network_interface.NI-TEST", &networkInterfaceId),
				),
			},
		},
	})

}

func testAccIfNetworkInterfaceExists(name string, networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		info, ok := stateInfo.RootModule().Resources[name]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceExists Failed,networkInterfaceName namely {%s} has not been created", info)
		}
		if info.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceExists Failed,operation failed,networkInterfaceName resources created but ID not set")
		}
		*networkInterfaceId = info.Primary.ID
		log.Printf("info.Primary.ID,", *networkInterfaceId)

		config := testAccProvider.Meta().(*JDCloudConfig)
		c := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeNetworkInterfaceRequest(config.Region, *networkInterfaceId)
		resp, err := c.DescribeNetworkInterface(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceExists Failed. Reasons:: code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		// ip + count
		l, _ := strconv.Atoi(info.Primary.Attributes["secondary_ip_addresses.#"])
		l2, _ := strconv.Atoi(info.Primary.Attributes["secondary_ip_count"])
		if l+l2 != len(resp.Result.NetworkInterface.SecondaryIps) {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceExists Failed,info don't mactch on secondary_ip_addresses.Details:"+
				"Local:%#v, Remote:%#v", l, resp.Result.NetworkInterface.SecondaryIps)
		}

		// sg
		l, _ = strconv.Atoi(info.Primary.Attributes["security_groups.#"])
		if l != len(resp.Result.NetworkInterface.NetworkSecurityGroupIds) {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceExists Failed,info don't mactch on security_groups.Details:"+
				"Local:%#v, Remote:%#v", info.Primary.Attributes["security_groups"], resp.Result.NetworkInterface.NetworkSecurityGroupIds)
		}

		// name
		if info.Primary.Attributes["network_interface_name"] != resp.Result.NetworkInterface.NetworkInterfaceName {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceExists Failed,info don't mactch on network_interface_name.Details:"+
				"Local:%#v, Remote:%#v", info.Primary.Attributes["network_interface_name"], resp.Result.NetworkInterface.NetworkInterfaceName)
		}

		return nil
	}
}

func testAccCheckNetworkInterfaceDestroy(networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *networkInterfaceId == "" {
			return fmt.Errorf("[ERROR] testAccCheckNetworkInterfaceDestroy Failed,networkInterfaceId is empty")
		}

		config := testAccProvider.Meta().(*JDCloudConfig)
		c := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeNetworkInterfaceRequest(config.Region, *networkInterfaceId)
		resp, err := c.DescribeNetworkInterface(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != RESOURCE_NOT_FOUND {
			return fmt.Errorf("[ERROR] testAccCheckNetworkInterfaceDestroy Failed,something wrong happens or resource still exists")
		}
		return nil
	}
}
