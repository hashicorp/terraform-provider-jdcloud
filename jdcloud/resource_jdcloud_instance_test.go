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

	TestCase : 1. Common stuff

			   2. [System-Disk][Pass] cn-north-1a only support local system disk with 40gb
								      So we are going to create in cn-east-1a , with cloud system disk,100gb
									  **By applying this test, change your "region" into cn-east-1

			   3. [Data-Disk][Fail] Usually we use cloud data disk, how about local data disk?
							  // *Unavailable since local-data-disk out of stock :-(

               4. [Image-ID][Pass] Create a vm with a private , customised image(rather than trivial Ubuntu stuff)

			   5. [Primary-IP] Create a vm without primary IP, they were supposed to be assigned a random one
			   6. [Security-Groups] Create a vm with random number of Security Groups
			   7. [Password] Create a vm without password
			   *8. Unexpected behavior for "instance_import" (currently removed)

*/

//1. Common stuff
const testAccInstanceGeneral = `
resource "jdcloud_instance" "terraform-normal-case" {
 az            = "cn-north-1a"
 instance_name = "%s"
 instance_type = "c.n1.large"
 image_id      = "img-chn8lfcn6j"
 password      = "%s"
 description   = "%s"

 subnet_id              = "subnet-j8jrei2981"
 network_interface_name = "jdcloud"
 primary_ip             = "10.0.5.0"
 security_group_ids     = ["sg-ym9yp1egi0"]

 elastic_ip_bandwidth_mbps = 10
 elastic_ip_provider       = "bgp"

 system_disk = {
   disk_category = "local"
   auto_delete   = true
   device_name   = "vda"
   disk_size_gb =  40
 }
}
`

func generateInstanceConfig(instanceName, password, description string) string {
	return fmt.Sprintf(testAccInstanceGeneral, instanceName, password, description)
}

func TestAccJDCloudInstance_basic(t *testing.T) {

	var instanceId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	name2 := randomStringWithLength(10)
	des2 := randomStringWithLength(20)

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_instance.terraform-normal-case",
		CheckDestroy:  testAccDiskInstanceDestroy("jdcloud_instance.terraform-normal-case", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: generateInstanceConfig(name1, "DevOps2018~", des1),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfInstanceExists(
						"jdcloud_instance.terraform-normal-case", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "instance_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "instance_type", "c.n1.large"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "image_id", "img-chn8lfcn6j"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "subnet_id", "subnet-j8jrei2981"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "network_interface_name", "jdcloud"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "primary_ip", "10.0.5.0"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "elastic_ip_provider", "bgp"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "system_disk.#", "1"),
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-normal-case", "data_disk"),
				),
			},
			{
				Config: generateInstanceConfig(name2, "DevOps2018!", des2),
				Check: resource.ComposeTestCheckFunc(
					testAccIfInstanceExists("jdcloud_instance.terraform-normal-case", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "instance_name", name2),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "instance_type", "c.n1.large"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "image_id", "img-chn8lfcn6j"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "description", des2),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "subnet_id", "subnet-j8jrei2981"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "network_interface_name", "jdcloud"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "primary_ip", "10.0.5.0"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "elastic_ip_provider", "bgp"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-normal-case", "system_disk.#", "1"),
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-normal-case", "data_disk"),
				),
			},
		},
	})
}

/*
   By starting this test, Change your region into cn-east-1
   Or error will be reported: Unable to find certain image_id

2. [System-Disk] Cloud-System-Disk on other cn area, larger size.
const testAccInstanceCloudSystemDisk = `
resource "jdcloud_instance" "terraform-cloud-sys-disk" {
 az            = "cn-east-2a"
 instance_name = "%s"
 instance_type = "g.n2.medium"
 image_id      = "%s"
 password      = "%s"
 description   = "%s"

 subnet_id              = "%s"
 security_group_ids     = ["%s"]
 elastic_ip_bandwidth_mbps = 10
 elastic_ip_provider       = "bgp"

 system_disk = {
   disk_category = "cloud"
	disk_type = "ssd.gp1"
   auto_delete   = true
   device_name   = "vda"
   disk_size_gb =  %s
 }
}
`

func instanceConfigCloudSysDisk(instanceName, image_id, password, description, subnet, sg, diskSize string) string {
	return fmt.Sprintf(testAccInstanceCloudSystemDisk, instanceName, image_id, password, description, subnet, sg, diskSize)
}

func TestAccJDCloudInstance_cloudSysDisk(t *testing.T) {

	var instanceId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	name2 := randomStringWithLength(10)
	des2 := randomStringWithLength(20)
	randSize := strconv.Itoa((rand.Intn(10) + 5) * 10)
	img_id := "img-2wul0u50re"
	subnet_id := "subnet-mu0sfxixbf"
	sg := "sg-ekogb6nv2b"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_instance.terraform-cloud-sys-disk",
		CheckDestroy:  testAccDiskInstanceDestroy("jdcloud_instance.terraform-cloud-sys-disk", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: instanceConfigCloudSysDisk(name1, img_id, "DevOps2018~", des1, subnet_id, sg, randSize),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfInstanceExists(
						"jdcloud_instance.terraform-cloud-sys-disk", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "az", "cn-east-2a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "instance_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-cloud-sys-disk", "ip_addresses.#"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.disk_category", "cloud"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.disk_type", "ssd.gp1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.disk_size_gb", randSize),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "data_disk"),
				),
			},
			{
				Config: instanceConfigCloudSysDisk(name2, img_id, "DevOps2018!", des2, subnet_id, sg, randSize),
				Check: resource.ComposeTestCheckFunc(
					testAccIfInstanceExists("jdcloud_instance.terraform-cloud-sys-disk", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "az", "cn-east-2a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "instance_name", name2),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "description", des2),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-cloud-sys-disk", "ip_addresses.#"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.disk_category", "cloud"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.disk_type", "ssd.gp1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.disk_size_gb", randSize),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-cloud-sys-disk", "data_disk"),
				),
			},
		},
	})
}
*/

// 4. [Image-ID] Create a vm with a private , customised image
const testAccInstancePrivateImage = `
resource "jdcloud_instance" "terraform-private-image" {
az            = "cn-north-1a"
instance_name = "%s"
instance_type = "g.n2.medium"
image_id      = "%s"
password      = "%s"
description   = "%s"

subnet_id              = "%s"
security_group_ids     = ["%s"]
elastic_ip_bandwidth_mbps = 10
elastic_ip_provider       = "bgp"

system_disk = {
  disk_category = "local"
  auto_delete   = true
  device_name   = "vda"
  disk_size_gb =  40
}
}
`

func instanceConfigPrivateImage(instanceName, image_id, password, description, subnet, sg string) string {
	return fmt.Sprintf(testAccInstancePrivateImage, instanceName, image_id, password, description, subnet, sg)
}

func TestAccJDCloudInstance_PrivateImage(t *testing.T) {

	var instanceId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	name2 := randomStringWithLength(10)
	des2 := randomStringWithLength(20)
	img_id := "img-c9hue6ckxd"
	subnet_id := "subnet-rht03mi6o0"
	sg := "sg-hzdy2lpzao"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_instance.terraform-private-image",
		CheckDestroy:  testAccDiskInstanceDestroy("jdcloud_instance.terraform-private-image", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: instanceConfigPrivateImage(name1, img_id, "DevOps2018~", des1, subnet_id, sg),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfInstanceExists(
						"jdcloud_instance.terraform-private-image", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "instance_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-private-image", "ip_addresses.#"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.disk_category", "local"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.disk_size_gb", "40"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-private-image", "data_disk"),
				),
			},
			{
				Config: instanceConfigPrivateImage(name2, img_id, "DevOps2018!", des2, subnet_id, sg),
				Check: resource.ComposeTestCheckFunc(
					testAccIfInstanceExists("jdcloud_instance.terraform-private-image", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "instance_name", name2),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "description", des2),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-private-image", "ip_addresses.#"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.disk_category", "local"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.disk_size_gb", "40"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-private-image", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-private-image", "data_disk"),
				),
			},
		},
	})
}

// 5. [Primary-IP] Create a vm without primary IP, they were supposed to be assigned a random one
const testAccInstancePrimaryIP = `
resource "jdcloud_instance" "terraform-primary-ip" {
az            = "cn-north-1a"
instance_name = "%s"
instance_type = "g.n2.medium"
image_id      = "%s"
password      = "%s"
description   = "%s"

subnet_id              = "%s"
security_group_ids     = ["%s"]
elastic_ip_bandwidth_mbps = 10
elastic_ip_provider       = "bgp"

system_disk = {
  disk_category = "local"
  auto_delete   = true
  device_name   = "vda"
  disk_size_gb =  40
}
}
`

func instanceConfigPrimaryIP(instanceName, image_id, password, description, subnet, sg string) string {
	return fmt.Sprintf(testAccInstancePrimaryIP, instanceName, image_id, password, description, subnet, sg)
}

func TestAccJDCloudInstance_PrimaryIP(t *testing.T) {

	var instanceId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	img_id := "img-chn8lfcn6j"
	subnet_id := "subnet-rht03mi6o0"
	sg := "sg-hzdy2lpzao"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_instance.terraform-primary-ip",
		CheckDestroy:  testAccDiskInstanceDestroy("jdcloud_instance.terraform-primary-ip", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: instanceConfigPrimaryIP(name1, img_id, "DevOps2018~", des1, subnet_id, sg),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfInstanceExists(
						"jdcloud_instance.terraform-primary-ip", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "instance_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-primary-ip", "ip_addresses.#"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "system_disk.0.disk_category", "local"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "system_disk.0.disk_size_gb", "40"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-primary-ip", "data_disk"),
				),
			},
		},
	})
}

// 6. [Security-Groups] Create a vm with random number of Security Groups
const testAccInstanceSG = `
resource "jdcloud_instance" "terraform-instance-sg" {
az            = "cn-north-1a"
instance_name = "%s"
instance_type = "g.n2.medium"
image_id      = "%s"
password      = "%s"
description   = "%s"

subnet_id              = "%s"
security_group_ids     = %s
elastic_ip_bandwidth_mbps = 10
elastic_ip_provider       = "bgp"

system_disk = {
  disk_category = "local"
  auto_delete   = true
  device_name   = "vda"
  disk_size_gb =  40
}
}
`

func instanceConfigSG(instanceName, image_id, password, description, subnet, sg string) string {
	return fmt.Sprintf(testAccInstanceSG, instanceName, image_id, password, description, subnet, sg)
}

func TestAccJDCloudInstanceSG(t *testing.T) {

	var instanceId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	img_id := "img-c9hue6ckxd"
	subnet_id := "subnet-rht03mi6o0"
	sg := `["sg-cl6uv4i782","sg-xmjw0695x0","sg-hzdy2lpzao"]`

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_instance.terraform-instance-sg",
		CheckDestroy:  testAccDiskInstanceDestroy("jdcloud_instance.terraform-instance-sg", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: instanceConfigSG(name1, img_id, "DevOps2018~", des1, subnet_id, sg),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfInstanceExists(
						"jdcloud_instance.terraform-instance-sg", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "instance_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-instance-sg", "ip_addresses.#"),

					// Validate on Security Groups
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-instance-sg", "security_group_ids.#"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "security_group_ids.#", "3"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "system_disk.0.disk_category", "local"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "system_disk.0.disk_size_gb", "40"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-instance-sg", "data_disk"),
				),
			},
		},
	})
}

// 7. [Password] Create a vm without password
const testAccInstancePassword = `
resource "jdcloud_instance" "terraform-instance-pw" {
 az            = "cn-north-1a"
 instance_name = "%s"
 instance_type = "g.n2.medium"
 image_id      = "%s"
 description   = "%s"

 subnet_id              = "%s"
 security_group_ids     = %s
 elastic_ip_bandwidth_mbps = 10
 elastic_ip_provider       = "bgp"

 system_disk = {
   disk_category = "local"
   auto_delete   = true
   device_name   = "vda"
   disk_size_gb =  40
 }
}
`

func instanceConfigPW(instanceName, image_id, description, subnet, sg string) string {
	return fmt.Sprintf(testAccInstancePassword, instanceName, image_id, description, subnet, sg)
}

func TestAccJDCloudInstancePW(t *testing.T) {

	var instanceId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	img_id := "img-c9hue6ckxd"
	subnet_id := "subnet-rht03mi6o0"
	sg := `["sg-cl6uv4i782","sg-xmjw0695x0","sg-hzdy2lpzao"]`

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_instance.terraform-instance-pw",
		CheckDestroy:  testAccDiskInstanceDestroy("jdcloud_instance.terraform-instance-pw", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: instanceConfigPW(name1, img_id, des1, subnet_id, sg),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfInstanceExists(
						"jdcloud_instance.terraform-instance-pw", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "instance_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-instance-pw", "ip_addresses.#"),

					// Validate on Security Groups
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-instance-pw", "security_group_ids.#"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "security_group_ids.#", "3"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "system_disk.0.disk_category", "local"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "system_disk.0.disk_size_gb", "40"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-instance-pw", "data_disk"),
				),
			},
		},
	})
}

// Currently, verification on disks is not available
func testAccIfInstanceExists(resourceName string, instanceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("testAccIfInstanceExists failed, resource %s unavailable", resourceName)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, resource:%s is created but ID not set", resourceName)
		}

		*instanceId = infoStoredLocally.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstanceRequest(config.Region, *instanceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return fmt.Errorf("testAccIfInstanceExists failed position-1")
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("testAccIfInstanceExists failed position-2")
		}

		return nil
	}
}

func testAccDiskInstanceDestroy(resourceName string, instanceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, _ := stateInfo.RootModule().Resources[resourceName]
		*instanceId = infoStoredLocally.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstanceRequest(config.Region, *instanceId)
		resp, err := vmClient.DescribeInstance(req)

		if err != nil {
			return fmt.Errorf("failed in deleting certain resources position-10")
		}

		if resp.Error.Code == REQUEST_COMPLETED {
			return fmt.Errorf("failed in deleting certain resources position-11 ,code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		return nil
	}
}
