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

const TestAccInstanceTemplateConfig = `
resource "jdcloud_instance_template" "instance_template" {
  "template_name" = "terraform_auto_change_name"
  "instance_type" = "g.n2.medium"
  "image_id" = "img-chn8lfcn6j"
  "password" = "DevOps2018"
  "bandwidth" = 5
  "ip_service_provider" = "BGP"
  "charge_mode" = "bandwith"
  "subnet_id" = "subnet-ge9rox69ul"
  "security_group_ids" = ["sg-eans7e93el"]
  "system_disk" = {
    disk_category = "local"
  }
  "data_disks" = {
    disk_category = "cloud"
  }
}
`

func TestAccJDCloudInstanceTemplate_basic(t *testing.T) {

	var instanceTemplateId string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccIfTemplateDestroyed(&instanceTemplateId),
		Steps: []resource.TestStep{
			{
				Config: TestAccInstanceTemplateConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfTemplateExists("jdcloud_instance_template.instance_template", &instanceTemplateId),
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
