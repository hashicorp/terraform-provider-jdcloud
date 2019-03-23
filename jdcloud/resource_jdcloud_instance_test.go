package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"math/rand"
	"strconv"
	"testing"
)

/*

	TestCase : 1. Common stuff
			   2. [System-Disk] cn-north-1a only support local system disk with 40gb
								So we are going to create in cn-east-1a , with cloud system disk,100gb
			   3. [Data-Disk][x] Usually we use cloud data disk, how about local data disk?
							  // *Unavailable since local-data-disk out of stock :-(
               4. [Image-ID] Create a vm with a private , customised image(rather than trivial Ubuntu stuff)
			   5. [Primary-IP] Create a vm without primary IP, they were supposed to be assigned a random one
			   6. [Secondary-IPs] Create a vm with random number of Secondary-IPs
			   7. [Security-Groups] Create a vm with random number of Security Groups
			   8. [Password] Create a vm without password

*/

// 1. Common stuff
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
			{
				ResourceName:      "jdcloud_instance.terraform-normal-case",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// 2. [System-Disk] Cloud-System-Disk on other cn area, larger size.
const testAccInstanceCloudSystemDisk = `
resource "jdcloud_instance" "terraform-cloud-sys-disk" {
  az            = "cn-east-1a"
  instance_name = "%s"
  instance_type = "g.n2.medium"
  image_id      = %s
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
	img_id := "img-chn8lfcn6j"
	subnet_id := "subnet-c41s6aa6n4"
	sg := "sg-ew56sb0km7"

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
						"jdcloud_instance.terraform-cloud-sys-disk", "az", "cn-north-1a"),
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
						"jdcloud_instance.terraform-cloud-sys-disk", "primary_ip"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-cloud-sys-disk", "ip_addresses"),

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
						"jdcloud_instance.terraform-cloud-sys-disk", "az", "cn-north-1a"),
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
						"jdcloud_instance.terraform-cloud-sys-disk", "primary_ip"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-cloud-sys-disk", "ip_addresses"),

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
				ResourceName:      "jdcloud_instance.terraform-cloud-sys-disk",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// 4. [Image-ID] Create a vm with a private , customised image
const testAccInstancePrivateImage = `
resource "jdcloud_instance" "terraform-private-image" {
  az            = "cn-north-1a"
  instance_name = "%s"
  instance_type = "g.n2.medium"
  image_id      = %s
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
						"jdcloud_instance.terraform-private-image", "primary_ip"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-private-image", "ip_addresses"),

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
						"jdcloud_instance.terraform-private-image", "primary_ip"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-private-image", "ip_addresses"),

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
				ResourceName:      "jdcloud_instance.terraform-private-image",
				ImportState:       true,
				ImportStateVerify: true,
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
  image_id      = %s
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
						"jdcloud_instance.terraform-primary-ip", "primary_ip"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-primary-ip", "ip_addresses"),

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
			{
				ResourceName:      "jdcloud_instance.terraform-primary-ip",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// 6. [Secondary-IPs] Create a vm with random number of Secondary-IPs
const testAccInstanceSecondaryIPs = `
resource "jdcloud_instance" "terraform-secondary-ips" {
  az            = "cn-north-1a"
  instance_name = "%s"
  instance_type = "g.n2.medium"
  image_id      = %s
  password      = "%s"
  description   = "%s"

  subnet_id              = "%s"
  security_group_ids     = ["%s"]
  elastic_ip_bandwidth_mbps = 10
  elastic_ip_provider       = "bgp"

  secondary_ips = %s
  secondary_ip_count = %s

  system_disk = {
    disk_category = "local"
    auto_delete   = true
    device_name   = "vda"
    disk_size_gb =  40
  }
}
`

func instanceConfigSecondaryIPs(instanceName, image_id, password, description, subnet, sg, secondary_ips, secondary_ip_count string) string {
	return fmt.Sprintf(testAccInstanceSecondaryIPs, instanceName, image_id, password, description, subnet, sg, secondary_ips, secondary_ip_count)
}

func TestAccJDCloudInstance_SecondaryIPs(t *testing.T) {

	var instanceId string
	name1 := randomStringWithLength(10)
	des1 := randomStringWithLength(20)
	img_id := "img-c9hue6ckxd"
	subnet_id := "subnet-rht03mi6o0"
	sg := "sg-hzdy2lpzao"
	secondary_ips := "[\"10.0.2.0\",\"10.0.3.0\"]"
	secondary_ip_count := "3"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		IDRefreshName: "jdcloud_instance.terraform-secondary-ips",
		CheckDestroy:  testAccDiskInstanceDestroy("jdcloud_instance.terraform-secondary-ips", &instanceId),
		Steps: []resource.TestStep{
			{
				Config: instanceConfigSecondaryIPs(name1, img_id, "DevOps2018~", des1, subnet_id, sg, secondary_ips, secondary_ip_count),
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfInstanceExists(
						"jdcloud_instance.terraform-secondary-ips", &instanceId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "az", "cn-north-1a"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "instance_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "image_id", img_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "description", des1),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "subnet_id", subnet_id),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "elastic_ip_bandwidth_mbps", "10"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "elastic_ip_provider", "bgp"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-secondary-ips", "primary_ip"),

					// Validate on Secondary IP addresses, there should be 5 in total
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-secondary-ips", "ip_addresses"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "ip_addresses.#", "5"),

					// Validate specs on system disk
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "system_disk.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "system_disk.0.disk_category", "local"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "system_disk.0.disk_size_gb", "40"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "system_disk.0.device_name", "vda"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "system_disk.0.auto_delete", "true"),

					// These values not supposed to exists after resource_XYZ_Read
					resource.TestCheckNoResourceAttr(
						"jdcloud_instance.terraform-secondary-ips", "data_disk"),
				),
			},
			{
				ResourceName:      "jdcloud_instance.terraform-secondary-ips",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// 7. [Security-Groups] Create a vm with random number of Security Groups
const testAccInstanceSG = `
resource "jdcloud_instance" "terraform-instance-sg" {
  az            = "cn-north-1a"
  instance_name = "%s"
  instance_type = "g.n2.medium"
  image_id      = %s
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
						"jdcloud_instance.terraform-instance-sg", "primary_ip"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-instance-sg", "ip_addresses"),

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
			{
				ResourceName:      "jdcloud_instance.terraform-instance-sg",
				ImportState:       true,
				ImportStateVerify: true,
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
  image_id      = %s
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
						"jdcloud_instance.terraform-instance-pw", "primary_ip"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance.terraform-instance-pw", "ip_addresses"),

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
			{
				ResourceName:      "jdcloud_instance.terraform-instance-pw",
				ImportState:       true,
				ImportStateVerify: true,
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

		localMap := infoStoredLocally.Primary.Attributes
		remoteStruct := resp.Result.Instance

		if remoteStruct.Description != localMap["description"] {
			return fmt.Errorf("testAccIfInstanceExists failed on description")
		}
		if remoteStruct.PrimaryNetworkInterface.NetworkInterface.PrimaryIp.PrivateIpAddress != localMap["primary_ip"] {
			return fmt.Errorf("testAccIfInstanceExists failed on primary ip")
		}
		if remoteStruct.ImageId != localMap["image_id"] {
			return fmt.Errorf("testAccIfInstanceExists failed on Image id")
		}
		if remoteStruct.InstanceName != localMap["instance_name"] {
			return fmt.Errorf("testAccIfInstanceExists failed on instance name ")
		}
		if remoteStruct.InstanceType != localMap["instance_type"] {
			return fmt.Errorf("testAccIfInstanceExists failed on instance type")
		}
		if remoteStruct.SubnetId != localMap["subnet_id"] {
			return fmt.Errorf("testAccIfInstanceExists failed subnet id")
		}
		if len(remoteStruct.KeyNames) != RESOURCE_EMPTY && remoteStruct.KeyNames[0] != localMap["key_names"] {
			return fmt.Errorf("testAccIfInstanceExists failed on key names")
		}
		sgLength, _ := strconv.Atoi(localMap["security_group_ids.#"])
		if len(remoteStruct.PrimaryNetworkInterface.NetworkInterface.SecurityGroups) != sgLength {
			return fmt.Errorf("testAccIfInstanceExists failed on security group")
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
