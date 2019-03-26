package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

/*
	TestCase : 1.common stuff , tricky point is to make sure the amount of
				secondary ips was as expected
			   -> Here, in the field, secondary IP count, you are only to add
*/
const TestAccNetWorkInterfaceTemplate = `
resource "jdcloud_network_interface" "terraform-ni"{
	subnet_id = "subnet-rht03mi6o0"
	description = "%s"
	az = "cn-north-1"
	network_interface_name = "%s"
	secondary_ip_addresses = %s 
	secondary_ip_count = %d
	security_groups = %s
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
				Config: generateNITemplate(
					"test", "TerraformTest", `["sg-hzdy2lpzao","sg-s0ardxmz3a"]`, `["10.0.3.0","10.0.4.0"]`, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccIfNetworkInterfaceExists(
						"jdcloud_network_interface.terraform-ni", &networkInterfaceId),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "subnet_id", "subnet-rht03mi6o0"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "description", "test"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "az", "cn-north-1"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "network_interface_name", "TerraformTest"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "security_groups.#", "2"),

					resource.TestCheckNoResourceAttr(
						"jdcloud_network_interface.terraform-ni", "primary_ip_address"),

					// Sanity check was by default set to 1
					resource.TestCheckResourceAttrSet(
						"jdcloud_network_interface.terraform-ni", "sanity_check"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "sanity_check", "1"),

					// By setting 2*ip_addr and 3*ip_count,we should get 5 ip_addresses in total
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "ip_addresses.#", "5"),
				),
			},
			{
				Config: generateNITemplate("BBCTopGear",
					"TerraformTestNewName",
					`["sg-hzdy2lpzao"]`,
					`["10.0.3.0"]`, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccIfNetworkInterfaceExists(
						"jdcloud_network_interface.terraform-ni", &networkInterfaceId),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "subnet_id", "subnet-rht03mi6o0"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "description", "BBCTopGear"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "az", "cn-north-1"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "network_interface_name", "TerraformTestNewName"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "security_groups.#", "1"),
					resource.TestCheckNoResourceAttr(
						"jdcloud_network_interface.terraform-ni", "primary_ip_address"),

					// Sanity check was by default set to 1
					resource.TestCheckResourceAttrSet(
						"jdcloud_network_interface.terraform-ni", "sanity_check"),
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "sanity_check", "1"),

					// After updating, we should get 6 ip_addresses here
					resource.TestCheckResourceAttr(
						"jdcloud_network_interface.terraform-ni", "ip_addresses.#", "6"),
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

func generateNITemplate(description, name, sg, ip_addr string, ip_count int) string {
	return fmt.Sprintf(TestAccNetWorkInterfaceTemplate, description, name, ip_addr, ip_count, sg)
}
