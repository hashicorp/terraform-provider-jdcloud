package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"testing"
)

/*

	TestCase : 1-[Pass]. Common stuff only. Not yet found any tricky point requires extra attention
			   2-[Pass]. Concurrent disk attachment/Detachment
*/

// 1-[Pass]. Common stuff only. Not yet found any tricky point requires extra attention
const TestAccDiskAttachmentTemplate = `
resource "jdcloud_disk_attachment" "terraform_da"{
	instance_id = "i-g6xse7qb0z" 
	disk_id = "vol-masm0gcxn8"
	auto_delete = %s
}
`

func diskAttachmentConfig(a string) string {
	return fmt.Sprintf(TestAccDiskAttachmentTemplate, a)
}

func TestAccJDCloudDiskAttachment_basic(t *testing.T) {

	var instanceId, diskId string

	resource.Test(t, resource.TestCase{

		IDRefreshName: "jdcloud_disk_attachment.terraform_da",
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccDiskAttachmentDestroy(&instanceId, &diskId),
		Steps: []resource.TestStep{
			{
				Config: diskAttachmentConfig("true"),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfDiskAttachmentExists(
						"jdcloud_disk_attachment.terraform_da", &instanceId, &diskId),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "instance_id", "i-g6xse7qb0z"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "disk_id", "vol-masm0gcxn8"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "auto_delete", "true"),

					// After resource_XYZ_Read these values will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_disk_attachment.terraform_da", "device_name"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "force_detach"),
				),
			},
			{
				Config: diskAttachmentConfig("false"),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfDiskAttachmentExists(
						"jdcloud_disk_attachment.terraform_da", &instanceId, &diskId),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "instance_id", "i-g6xse7qb0z"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "disk_id", "vol-masm0gcxn8"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "auto_delete", "false"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_disk_attachment.terraform_da", "device_name"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_disk_attachment.terraform_da", "force_detach"),
				),
			},
		},
	})
}

// 2-[Pass]. Concurrent disk attachment/Detachment
const TestAccDiskAttachmentConcurrentAttach = `
resource "jdcloud_disk_attachment" "terraform_con_1"{
	instance_id = "i-g6xse7qb0z" 
	disk_id = "vol-qm7t7q7pmk"
}
resource "jdcloud_disk_attachment" "terraform_con_2"{
	instance_id = "i-g6xse7qb0z" 
	disk_id = "vol-9dya7e5rdi"
}
`

func TestAccJDCloudDiskAttachment_concurrent_attach(t *testing.T) {

	var instanceId, diskId string
	var instanceId2, diskId2 string

	resource.Test(t, resource.TestCase{

		IDRefreshName: "jdcloud_disk_attachment.terraform_da",
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccDiskAttachmentBothDestroy(&instanceId, &diskId, &diskId2),
		Steps: []resource.TestStep{
			{
				Config: TestAccDiskAttachmentConcurrentAttach,
				Check: resource.ComposeTestCheckFunc(

					// Both of them are supposed to exist
					testAccIfDiskAttachmentExists(
						"jdcloud_disk_attachment.terraform_con_1", &instanceId, &diskId),
					testAccIfDiskAttachmentExists(
						"jdcloud_disk_attachment.terraform_con_2", &instanceId2, &diskId2),

					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_con_1", "instance_id", "i-g6xse7qb0z"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_con_2", "instance_id", "i-g6xse7qb0z"),

					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_con_1", "disk_id", "vol-qm7t7q7pmk"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk_attachment.terraform_con_2", "disk_id", "vol-9dya7e5rdi"),

					// After resource_XYZ_Read these values will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_disk_attachment.terraform_con_1", "device_name"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_disk_attachment.terraform_con_2", "device_name"),
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

func testAccDiskAttachmentBothDestroy(resourceId *string, diskId, diskId2 *string) resource.TestCheckFunc {

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
			if *diskId2 == disk.CloudDisk.DiskId && disk.Status != DISK_DETACHED {
				return fmt.Errorf("[ERROR] testAccDiskAttachmentDestroy failed,data disk failed in detatching")
			}
		}
		return nil
	}
}
