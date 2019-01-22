package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	"strconv"
	"testing"
)

const TestAccDiskConfig = `
resource "jdcloud_disk" "disk_test_1" {
  az           = "cn-north-1a"
  name         = "test_disk"
  description  = "test123"
  disk_type    = "ssd"
  disk_size_gb = 20
  charge_mode = "postpaid_by_duration"
}
`

func TestAccJDCloudDisk_basic(t *testing.T) {

	var diskId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDiskDestroy(&diskId),
		Steps: []resource.TestStep{
			{
				Config: TestAccDiskConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfDiskExists("jdcloud_disk.disk_test_1", &diskId),
				),
			},
		},
	})
}

func testAccIfDiskExists(diskName string, diskId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		localDiskInfo, ok := stateInfo.RootModule().Resources[diskName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfDiskExists failed, we can not find a disk namely:{%s} in terraform.State", diskName)
		}
		if localDiskInfo.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfDiskExists failed,operation failed, Disk is created but ID not set")
		}
		*diskId = localDiskInfo.Primary.ID

		diskConfig := testAccProvider.Meta().(*JDCloudConfig)
		diskClient := client.NewDiskClient(diskConfig.Credential)

		req := apis.NewDescribeDiskRequest(diskConfig.Region, *diskId)
		resp, err := diskClient.DescribeDisk(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfDiskExists failed according to the ID stored locally,we cannot find any RouteTable created remotely")
		}
		if localDiskInfo.Primary.Attributes["az"] != resp.Result.Disk.Az {
			return fmt.Errorf("[ERROR] testAccIfDiskExists failed info does not match on az")
		}
		if localDiskInfo.Primary.Attributes["disk_size_gb"] != strconv.Itoa(resp.Result.Disk.DiskSizeGB) {
			return fmt.Errorf("[ERROR] testAccIfDiskExists failed info does not match on disk_size_gb")
		}
		if localDiskInfo.Primary.Attributes["disk_type"] != resp.Result.Disk.DiskType {
			return fmt.Errorf("[ERROR] testAccIfDiskExists failed info does not match on disktype")
		}
		if localDiskInfo.Primary.Attributes["name"] != resp.Result.Disk.Name {
			return fmt.Errorf("[ERROR] testAccIfDiskExists failed info does not match on name")
		}

		return nil
	}
}

func testAccCheckDiskDestroy(diskId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *diskId == "" {
			return errors.New("[ERROR] testAccCheckDiskDestroy Failed subnetID is empty")
		}

		diskConfig := testAccProvider.Meta().(*JDCloudConfig)
		diskClient := client.NewDiskClient(diskConfig.Credential)

		req := apis.NewDescribeDiskRequest(diskConfig.Region, *diskId)
		resp, err := diskClient.DescribeDisk(req)

		if err != nil {
			return err
		}
		if resp.Result.Disk.Status != DISK_DELETED {
			return fmt.Errorf("[ERROR] testAccCheckDiskDestroy Failed, resource still exists DiskId: %s, DiskStatus: %s", *diskId, resp.Result.Disk.Status)
		}

		return nil
	}
}
