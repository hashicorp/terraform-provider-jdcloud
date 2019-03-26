package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

const TestAccNetworkInterfaceAttachmentConfig = `
resource "jdcloud_network_interface_attachment" "attachment-TEST-1"{
  instance_id = "i-g6xse7qb0z"
  network_interface_id = "port-4nwidjolb3"
}
`

func TestAccJDCloudNetworkInterfaceAttachment_basic(t *testing.T) {

	var networkInterfaceId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkInterfaceAttachmentDestroy(&networkInterfaceId),
		Steps: []resource.TestStep{
			{
				Config: TestAccNetworkInterfaceAttachmentConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfNetworkInterfaceAttachmentExists("jdcloud_network_interface_attachment.attachment-TEST-1", &networkInterfaceId),
					resource.TestCheckResourceAttr("jdcloud_network_interface_attachment.attachment-TEST-1", "instance_id", "i-g6xse7qb0z"),
					resource.TestCheckResourceAttr("jdcloud_network_interface_attachment.attachment-TEST-1", "network_interface_id", "port-4nwidjolb3"),

					// auto_delete shouldn't be here since they were not set in resource_XYZ_Read
					resource.TestCheckNoResourceAttr("jdcloud_network_interface_attachment.attachment-TEST-1", "auto_delete"),
				),
			},
		},
	})

}

func testAccIfNetworkInterfaceAttachmentExists(attachmentName string, networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		info, ok := stateInfo.RootModule().Resources[attachmentName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.attachment namely {%s} has not been created", attachmentName)
		}

		*networkInterfaceId, ok = info.Primary.Attributes["network_interface_id"]
		if info.Primary.ID == "" || ok == false {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.operation failed, resources created but ID not set")
		}

		config := testAccProvider.Meta().(*JDCloudConfig)
		c := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeNetworkInterfaceRequest(config.Region, *networkInterfaceId)
		resp, err := c.DescribeNetworkInterface(req)

		if err != nil {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.Create check  failed ,error message: %s", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.resources created locally but not remotely")
		}

		//instanceIdLocal := info.Primary.Attributes["instance_id"]
		//instanceIdRemote := resp.Result.NetworkInterface.InstanceId
		//
		//if instanceIdLocal != instanceIdRemote {
		//	return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.does not match, local:%s remote:%s",instanceIdLocal,instanceIdRemote)
		//}

		return nil
	}
}

func testAccCheckNetworkInterfaceAttachmentDestroy(networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *networkInterfaceId == "" {
			return fmt.Errorf("[ERROR] testAccCheckNetworkInterfaceAttachmentDestroy Failed.networkInterfaceId is empty")
		}

		attachmentConfig := testAccProvider.Meta().(*JDCloudConfig)
		attachmentClient := client.NewVpcClient(attachmentConfig.Credential)

		req := apis.NewDescribeNetworkInterfaceRequest(attachmentConfig.Region, *networkInterfaceId)
		resp, err := attachmentClient.DescribeNetworkInterface(req)

		if err != nil {
			return fmt.Errorf("[ERROR] testAccCheckNetworkInterfaceAttachmentDestroy Failed.delete check  failed ,error message: %s", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccCheckNetworkInterfaceAttachmentDestroy Failed.something wrong happens or resource still exists")
		}

		if resp.Result.NetworkInterface.InstanceId != "" {
			return fmt.Errorf("[ERROR] testAccCheckNetworkInterfaceAttachmentDestroy Failed.failed %s", resp.Result.NetworkInterface.InstanceId)
		}
		return nil
	}
}
