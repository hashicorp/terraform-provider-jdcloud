package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/pkg/errors"
	"testing"
)

const TestAccNetworkInterfaceAttachmentConfig = `
resource "jdcloud_network_interface_attachment" "attachment-TEST-1"{
  instance_id = "i-hves6944st"
  network_interface_id = "port-ampj4oamxw"
  auto_delete = "true"
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
				),
			},
		},
	})

}

func testAccIfNetworkInterfaceAttachmentExists(attachmentName string, networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		attachmentInfoStoredLocally, ok := stateInfo.RootModule().Resources[attachmentName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.attachment namely {%s} has not been created", attachmentName)
		}

		networkInterfaceIdLocal, ok := attachmentInfoStoredLocally.Primary.Attributes["network_interface_id"]
		if attachmentInfoStoredLocally.Primary.ID == "" || ok == false {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.operation failed, resources created but ID not set")
		}

		attachmentConfig := testAccProvider.Meta().(*JDCloudConfig)
		attachmentClient := client.NewVpcClient(attachmentConfig.Credential)

		req := apis.NewDescribeNetworkInterfaceRequest(attachmentConfig.Region, networkInterfaceIdLocal)
		resp, err := attachmentClient.DescribeNetworkInterface(req)

		if err != nil {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.Create check  failed ,error message: %s", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.resources created locally but not remotely")
		}

		instanceIdLocal := attachmentInfoStoredLocally.Primary.Attributes["instance_id"]
		instanceIdRemote := resp.Result.NetworkInterface.InstanceId

		if instanceIdLocal != instanceIdRemote {
			return fmt.Errorf("[ERROR] testAccIfNetworkInterfaceAttachmentExists Failed.resources locally and remotely does not match")
		}

		*networkInterfaceId = networkInterfaceIdLocal
		return nil
	}
}

func testAccCheckNetworkInterfaceAttachmentDestroy(networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *networkInterfaceId == "" {
			return errors.New("[ERROR] testAccCheckNetworkInterfaceAttachmentDestroy Failed.networkInterfaceId is empty")
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
