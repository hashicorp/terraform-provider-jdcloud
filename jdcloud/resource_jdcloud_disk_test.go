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
	"time"
)

const TestAccDiskConfig = `
resource "jdcloud_disk" "disk_test_1" {
  az           = "cn-north-1a"
  name         = "test_disk"
  description  = "test"
  disk_type    = "premium-hdd"
  disk_size_gb = 50
}
`

func TestAccJDCloudDisk_basic(t *testing.T){

	var diskId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckDiskDestroy(&diskId),
		Steps: []resource.TestStep{
			{
				Config: TestAccDiskConfig,
				Check: resource.ComposeTestCheckFunc(

					// SUBNET_ID verification
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
			return fmt.Errorf("we can not find a disk namely:{%s} in terraform.State", diskName)
		}
		if localDiskInfo.Primary.ID == "" {
			return fmt.Errorf("operation failed, Disk is created but ID not set")
		}
		*diskId = localDiskInfo.Primary.ID

		diskConfig := testAccProvider.Meta().(*JDCloudConfig)
		diskClient := client.NewDiskClient(diskConfig.Credential)

		req := apis.NewDescribeDiskRequest(diskConfig.Region,*diskId)
		resp, err := diskClient.DescribeDisk(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("according to the ID stored locally,we cannot find any RouteTable created remotely")
		}
		if localDiskInfo.Primary.Attributes["az"]!=resp.Result.Disk.Az {
			return fmt.Errorf("info does not match on az")
		}
		if localDiskInfo.Primary.Attributes["disk_size_gb"]!= strconv.Itoa(resp.Result.Disk.DiskSizeGB) {
			return fmt.Errorf("info does not match on disk_size_gb")
		}
		if localDiskInfo.Primary.Attributes["disk_type"]!=resp.Result.Disk.DiskType {
			return fmt.Errorf("info does not match on disktype")
		}
		if localDiskInfo.Primary.Attributes["name"]!=resp.Result.Disk.Name {
			return fmt.Errorf("info does not match on name")
		}

		return nil
	}
}

func testAccCheckDiskDestroy(diskId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {


		if*diskId=="" {
			return errors.New("subnetID is empty")
		}

		diskConfig := testAccProvider.Meta().(*JDCloudConfig)
		diskClient := client.NewDiskClient(diskConfig.Credential)

		req := apis.NewDescribeDiskRequest(diskConfig.Region,*diskId)

		retryCount := 0
		retryTag:
			resp, err := diskClient.DescribeDisk(req)

		if err!=nil {
			return err
		}

		if resp.Result.Disk.Status == "deleting" && retryCount<3{
			retryCount++
			time.Sleep(time.Second*3)
			goto retryTag
		}

		if resp.Result.Disk.Status!="deleted"{
			return fmt.Errorf("resource still exists %s,%s",*diskId,resp.Result.Disk.Status)
		}
		return nil
	}
}
