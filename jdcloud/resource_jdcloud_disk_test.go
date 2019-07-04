package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	"math/rand"
	"strconv"
	"testing"
)

/*
	TestCase : 1-[Pass]. Common stuff
			   2-[Pass]. ChargeMode are supposed to be "postpaid_by_duration" by default if not given
               3-[Fail]. Not sure if "postpaid_by_usage" chargeMode are available thus requires an exXtra test
				 Latest updates : postpaid_by_usage unavailable for disk, avoid it.
				[Discarded] "Postpaid_By_Usage" has been discarded

			   4-[Pass]. For an empty filed e.g snapshot id, It's expected to be "" && NotSet
*/

const (
	testAccDiskTemplateNoChargeMode = `
resource "jdcloud_disk" "terraform_dt_nc" {
  az           = "cn-north-1a"
  name         = "%s"
  description  = "%s"
  disk_type    = "ssd"
  disk_size_gb = %s
}
`
)

func generateDiskConfigNoChargeMode(name, description, diskSize string) string {
	return fmt.Sprintf(testAccDiskTemplateNoChargeMode, name, description, diskSize)
}

// Default payment type
func TestAccJDCloudDisk_default_charge_mode(t *testing.T) {

	var diskId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	name2 := randomStringWithLength(10)
	des2 := randomStringWithLength(20)
	randSize := strconv.Itoa((rand.Intn(10) + 5) * 10)

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_disk.terraform_dt_nc",
		CheckDestroy:  testAccCheckDiskDestroy(&diskId),

		Steps: []resource.TestStep{
			{
				Config: generateDiskConfigNoChargeMode(name1, des1, randSize),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfDiskExists("jdcloud_disk.terraform_dt_nc", &diskId),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "disk_type", "ssd"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "disk_size_gb", randSize),

					// After resource_XYZ_Read these value will be set to a certain value
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "snapshot_id", ""),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckResourceAttrSet(
						"jdcloud_disk.terraform_dt_nc", "charge_mode"),
					resource.TestCheckNoResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "charge_duration"),
					resource.TestCheckNoResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "charge_unit"),
				),
			},
			{
				Config: generateDiskConfigNoChargeMode(name2, des2, randSize),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfDiskExists("jdcloud_disk.terraform_dt_nc", &diskId),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "name", name2),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "description", des2),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "disk_type", "ssd"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "disk_size_gb", randSize),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "charge_mode", "postpaid_by_duration"),

					// After resource_XYZ_Read these value will be set to a certain value
					resource.TestCheckResourceAttrSet(
						"jdcloud_disk.terraform_dt_nc", "charge_mode"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "charge_mode", "postpaid_by_duration"),
					resource.TestCheckResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "snapshot_id", ""),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "charge_duration"),
					resource.TestCheckNoResourceAttr(
						"jdcloud_disk.terraform_dt_nc", "charge_unit"),
				),
			},
			{
				ResourceName:      "jdcloud_disk.terraform_dt_nc",
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
