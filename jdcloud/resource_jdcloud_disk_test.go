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

const (
	testAccDiskTemplate = `
resource "jdcloud_disk" "terraform_disk_test" {
  az           = "cn-north-1a"
  name         = "%s"
  description  = "%s"
  disk_type    = "ssd"
  disk_size_gb = 20
  charge_mode = "postpaid_by_duration"
}
`
)

// This function test -> `DiskCreate` & `DiskUpdate`
func TestAccJDCloudDisk_basic(t *testing.T) {

	var diskId string

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_disk.terraform_disk_test",
		CheckDestroy:  testAccCheckDiskDestroy(&diskId),

		Steps: []resource.TestStep{
			{
				Config: generateDiskConfig("a_normal_disk", "Auto generated normal disk, nothing special"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfDiskExists("jdcloud_disk.terraform_disk_test", &diskId),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "name", "a_normal_disk"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "description", "Auto generated normal disk, nothing special"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "disk_type", "ssd"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "disk_size_gb", "20"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "charge_mode", "postpaid_by_duration"),
					resource.TestCheckNoResourceAttr("jdcloud_disk.terraform_disk_test", "snapshot_id"),
					resource.TestCheckNoResourceAttr("jdcloud_disk.terraform_disk_test", "charge_duration"),
					resource.TestCheckNoResourceAttr("jdcloud_disk.terraform_disk_test", "charge_unit"),
				),
			},
			{
				Config: generateDiskConfig("normal_disk_with_new_name", "Still the same one, just different name"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfDiskExists("jdcloud_disk.terraform_disk_test", &diskId),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "name", "normal_disk_with_new_name"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "description", "Still the same one, just different name"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "disk_type", "ssd"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "disk_size_gb", "20"),
					resource.TestCheckResourceAttr("jdcloud_disk.terraform_disk_test", "charge_mode", "postpaid_by_duration"),
					resource.TestCheckNoResourceAttr("jdcloud_disk.terraform_disk_test", "snapshot_id"),
					resource.TestCheckNoResourceAttr("jdcloud_disk.terraform_disk_test", "charge_duration"),
					resource.TestCheckNoResourceAttr("jdcloud_disk.terraform_disk_test", "charge_unit"),
				),
			},
			{
				ResourceName:      "jdcloud_disk.terraform_disk_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccIfDiskExists(diskName string, diskId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		localDiskInfo, ok := stateInfo.RootModule().Resources[diskName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfDiskExists, DiskNotFound namely {%s} not found ", diskName)
		}
		if localDiskInfo.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfDiskExists, Disk is created but ID not set")
		}
		*diskId = localDiskInfo.Primary.ID

		diskConfig := testAccProvider.Meta().(*JDCloudConfig)
		diskClient := client.NewDiskClient(diskConfig.Credential)

		req := apis.NewDescribeDiskRequest(diskConfig.Region, *diskId)
		resp, err := diskClient.DescribeDisk(req)
		fmt.Printf("creating-%v", resp.Result.Disk)
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

func generateDiskConfig(name, description string) string {
	return fmt.Sprintf(testAccDiskTemplate, name, description)
}
