package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"testing"
)

const TestAccDiskAttachmentConfig = `
resource "jdcloud_disk_attachment" "disk-attachment-TEST-1"{
	instance_id = "i-g6xse7qb0z" 
	disk_id = "vol-w7bhp8s43l"
}
`

func TestAccJDCloudDiskAttachment_basic(t *testing.T) {

	var instanceId, diskId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDiskAttachmentDestroy(&instanceId, &diskId),
		Steps: []resource.TestStep{
			{
				Config: TestAccDiskAttachmentConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfDiskAttachmentExists("jdcloud_disk_attachment.disk-attachment-TEST-1", &instanceId, &diskId),
				),
			},
		},
	})
}

func testAccIfDiskAttachmentExists(resourceName string, resourceId, diskId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfDiskAttachmentExists failed, we can not find a resouce namely:{%s} in terraform.State", resourceName)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfDiskAttachmentExists failed, operation failed, resource:%s is created but ID not set", resourceName)
		}
		*resourceId = infoStoredLocally.Primary.Attributes["instance_id"]
		*diskId = infoStoredLocally.Primary.Attributes["disk_id"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)

		req := apis.NewDescribeInstanceRequest(config.Region, *resourceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return err
		}

		for _, aDisk := range resp.Result.Instance.DataDisks {

			if aDisk.CloudDisk.DiskId == *diskId {

				return nil
			}
		}

		return fmt.Errorf("[ERROR] testAccIfDiskAttachmentExists failed,resource not found remotely")
	}
}

func testAccDiskAttachmentDestroy(resourceId *string, diskId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)

		req := apis.NewDescribeInstanceRequest(config.Region, *resourceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return err
		}

		for _, disk := range resp.Result.Instance.DataDisks {
			if *diskId == disk.CloudDisk.DiskId && disk.Status != DISK_DETACHED {
				return fmt.Errorf("[ERROR] testAccDiskAttachmentDestroy failed,data disk failed in detatching")
			}
		}
		return nil
	}
}
