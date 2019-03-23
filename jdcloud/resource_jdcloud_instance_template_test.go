package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"strconv"
	"testing"
	"time"
)

/*
	TestCase : 1.common stuff
			   2. [Data-Disk] Build an instance template with multiple same data-disks
			   3. [EIP] REALLY annoying, create one with EIP and one without EIP
*/
// 1 Common stuff (Without EIP)
const TestAccInstanceTemplateTemplate = `
resource "jdcloud_instance_template" "instance_template" {
  "template_name" = "%s"
  "instance_type" = "g.n2.medium"
  "image_id" = "img-chn8lfcn6j"
  "password" = "DevOps2018"
  "subnet_id" = "subnet-rht03mi6o0"
  "security_group_ids" = ["sg-hzdy2lpzao"]
  "system_disk" = {
    disk_category = "local"
  }
  "data_disks" = {
    disk_category = "cloud"
  }
}
`

func generateInstanceTemplate(name string) string {
	return fmt.Sprintf(TestAccInstanceTemplateTemplate, name)
}

func TestAccJDCloudInstanceTemplate_basic(t *testing.T) {

	var instanceTemplateId string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccIfTemplateDestroyed(&instanceTemplateId),
		Steps: []resource.TestStep{
			{
				Config: generateInstanceTemplate("terraform_auto_change_name"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfTemplateExists("jdcloud_instance_template.instance_template", &instanceTemplateId),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "template_name", "terraform_auto_change_name"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "image_id", "img-chn8lfcn6j"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "subnet_id", "subnet-rht03mi6o0"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "system_disk.#", "1"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "data_disks.#", "1"),
					resource.TestCheckResourceAttrSet("jdcloud_instance_template.instance_template", "data_disks.0.device_name"),
				),
			},
			{
				Config: generateInstanceTemplate("another_name"),
				Check: resource.ComposeTestCheckFunc(
					testAccIfTemplateExists("jdcloud_instance_template.instance_template", &instanceTemplateId),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "template_name", "another_name"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "image_id", "img-chn8lfcn6j"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "subnet_id", "subnet-rht03mi6o0"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "system_disk.#", "1"),
					resource.TestCheckResourceAttr("jdcloud_instance_template.instance_template", "data_disks.#", "1"),
					resource.TestCheckResourceAttrSet("jdcloud_instance_template.instance_template", "data_disks.0.device_name"),
				),
			},
		},
	})
}

//2. Build an instance template with multiple same data-disks (Without EIP)
const TestAccInstanceTemplateMultipleDisk = `
resource "jdcloud_instance_template" "instance_template_md" {
 "template_name" = "%s"
 "instance_type" = "g.n2.medium"
 "image_id" = "img-chn8lfcn6j"
 "password" = "DevOps2018"
 "subnet_id" = "subnet-rht03mi6o0"
 "security_group_ids" = ["sg-hzdy2lpzao"]
 "system_disk" = {
   disk_category = "local"
 }
 "data_disks" = [
 {
   disk_category = "cloud"
 },
 {
   disk_category = "cloud"
 },
 {
   disk_category = "cloud"
 }]
}
`

func instanceTemplateMD(name string) string {
	return fmt.Sprintf(TestAccInstanceTemplateMultipleDisk, name)
}

func TestAccJDCloudInstanceTemplate_MultipleDisk(t *testing.T) {

	var instanceTemplateId string
	name := randomStringWithLength(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccIfTemplateDestroyed(&instanceTemplateId),
		Steps: []resource.TestStep{
			{
				Config: instanceTemplateMD(name),
				Check: resource.ComposeTestCheckFunc(
					testAccIfTemplateExists(
						"jdcloud_instance_template.instance_template_md", &instanceTemplateId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_md", "template_name", name),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_md", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_md", "image_id", "img-chn8lfcn6j"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_md", "subnet_id", "subnet-rht03mi6o0"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_md", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_md", "system_disk.#", "1"),

					// Validate on DataDisks
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_md", "data_disks.#", "3"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.0.device_name"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.0.disk_size"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.0.disk_type"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.0.disk_category"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.0.auto_delete"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.1.device_name"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.1.disk_size"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.1.disk_type"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.1.disk_category"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.1.auto_delete"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.2.device_name"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.2.disk_size"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.2.disk_type"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.2.disk_category"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.2.auto_delete"),
					resource.TestCheckResourceAttrSet(
						"jdcloud_instance_template.instance_template_md", "data_disks.2.device_name"),
				),
			},
		},
	})
}

//3. [EIP] - Create one with EIP
const TestAccInstanceTemplateBandwidth = `
resource "jdcloud_instance_template" "instance_template_bandwidth" {
 "template_name" = "%s"
 "instance_type" = "g.n2.medium"
 "image_id" = "img-chn8lfcn6j"
 "password" = "DevOps2018"
 "bandwidth" = 5
 "ip_service_provider" = "BGP"
 "charge_mode" = "bandwith"
 "subnet_id" = "subnet-rht03mi6o0"
 "security_group_ids" = ["sg-hzdy2lpzao"]
 "system_disk" = {
   disk_category = "local"
 }
 "data_disks" = [
 {
   disk_category = "cloud"
 }]
}
`

func instanceTemplateBandwidth(name string) string {
	return fmt.Sprintf(TestAccInstanceTemplateBandwidth, name)
}

func TestAccJDCloudInstanceTemplate_Bandwidth(t *testing.T) {

	var instanceTemplateId string
	name := randomStringWithLength(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccIfTemplateDestroyed(&instanceTemplateId),
		Steps: []resource.TestStep{
			{
				Config: instanceTemplateBandwidth(name),
				Check: resource.ComposeTestCheckFunc(
					testAccIfTemplateExists(
						"jdcloud_instance_template.instance_template_bandwidth", &instanceTemplateId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "template_name", name),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "instance_type", "g.n2.medium"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "image_id", "img-chn8lfcn6j"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "subnet_id", "subnet-rht03mi6o0"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "system_disk.#", "1"),

					// Validate EIP stuff
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "ip_service_provider", "BGP"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "charge_mode", "bandwith"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "bandwidth", "5"),

					// Validate on DataDisks
					resource.TestCheckResourceAttr(
						"jdcloud_instance_template.instance_template_bandwidth", "data_disks.#", "1"),
				),
			},
		},
	})
}

func testAccIfTemplateExists(templateName string, templateId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		localTemplateInfo, ok := stateInfo.RootModule().Resources[templateName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfTemplateExists failed, we can not find a template namely:{%s} in terraform.State", templateName)
		}
		if localTemplateInfo.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfTemplateExists failed,operation failed, template is created but ID not set")
		}

		*templateId = localTemplateInfo.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)

		req := apis.NewDescribeInstanceTemplateRequest(config.Region, *templateId)
		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.DescribeInstanceTemplate(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {

				if resp.Result.InstanceTemplate.Name != localTemplateInfo.Primary.Attributes["template_name"] {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.template_name(%s) != remote template_name(%s)", localTemplateInfo.Primary.Attributes["availability_group_name"], resp.Result.InstanceTemplate.Name))
				}

				if resp.Result.InstanceTemplate.InstanceTemplateData.InstanceType != localTemplateInfo.Primary.Attributes["instance_type"] {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.instance_type(%s) != remote instance_type(%s)", localTemplateInfo.Primary.Attributes["instance_type"], resp.Result.InstanceTemplate.InstanceTemplateData.InstanceType))
				}

				if resp.Result.InstanceTemplate.InstanceTemplateData.ImageId != localTemplateInfo.Primary.Attributes["image_id"] {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.image_id(%s) != remote image_id(%s)", localTemplateInfo.Primary.Attributes["image_id"], resp.Result.InstanceTemplate.InstanceTemplateData.ImageId))
				}

				if resp.Result.InstanceTemplate.InstanceTemplateData.PrimaryNetworkInterface.NetworkInterface.SubnetId != localTemplateInfo.Primary.Attributes["subnet_id"] {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.subnet_id(%s) != remote subnet_id(%s)", localTemplateInfo.Primary.Attributes["availability_group_name"], resp.Result.InstanceTemplate.InstanceTemplateData.PrimaryNetworkInterface.NetworkInterface.SubnetId))
				}

				localSgLength, _ := strconv.Atoi(localTemplateInfo.Primary.Attributes["security_group_ids"])
				if len(resp.Result.InstanceTemplate.InstanceTemplateData.PrimaryNetworkInterface.NetworkInterface.SecurityGroups) != localSgLength {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.security_group_ids(%d) != remote security_group_ids(%d)", localSgLength, len(resp.Result.InstanceTemplate.InstanceTemplateData.PrimaryNetworkInterface.NetworkInterface.SecurityGroups)))
				}

				localDiskDataLength, _ := strconv.Atoi(localTemplateInfo.Primary.Attributes["data_disks"])
				if len(resp.Result.InstanceTemplate.InstanceTemplateData.DataDisks) != localDiskDataLength {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.data_disks(%d) != remote data_disks(%d)", localDiskDataLength, len(resp.Result.InstanceTemplate.InstanceTemplateData.DataDisks)))
				}
				return nil
			}

			if connectionError(err) {
				return resource.RetryableError(formatConnectionErrorMessage())
			} else {
				return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
			}
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccIfTemplateDestroyed(templateId *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {

		if *templateId == "" {
			return fmt.Errorf("[ERROR] testAccIfTemplateDestroyed Failed templateId is empty")
		}

		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstanceTemplateRequest(config.Region, *templateId)

		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.DescribeInstanceTemplate(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				return resource.NonRetryableError(fmt.Errorf("[E] testAccIfTemplateDestroyed failed, resource still exists"))
			}

			if resp.Error.Code == RESOURCE_NOT_FOUND {
				return nil
			}

			if connectionError(err) {
				return resource.RetryableError(formatConnectionErrorMessage())
			} else {
				return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
			}
		})

		if err != nil {
			return err
		}
		return nil
	}
}
