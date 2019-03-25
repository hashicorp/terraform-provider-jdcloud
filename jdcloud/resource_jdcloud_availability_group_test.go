package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/ag/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/ag/client"
	"strconv"
	"testing"
	"time"
)

/*
	TestCase : 1.common stuff only. Not yet found any tricky point requires extra attention
*/

func agConfigSingleAz(name, description string) string {
	return fmt.Sprintf(testAccAGTemplate, name, description)
}

const testAccAGTemplate = `
resource "jdcloud_availability_group" "terraform_ag" {
  availability_group_name = "%s"
  az = ["cn-north-1a"]
  instance_template_id = "it-fpxvfdch26"
  description  = "%s"
  ag_type = "kvm"
}
`

// Multi-Az requires them to be in the same location ->> cn-north-1a & cn-north-1b
//                                                   ->> cn-east-1a & cn-east-1b
func agConfigDualAz(name string) string {
	return fmt.Sprintf(testAccAGTemplateDualAzTemplate, name)
}

const testAccAGTemplateDualAzTemplate = `
resource "jdcloud_availability_group" "terraform_ag_daz" {
  availability_group_name = "%s"
  az = ["cn-north-1a","cn-north-1b"]
  instance_template_id = "it-fpxvfdch26"
  ag_type = "docker"
}
`

// common case ->> Create + Update
func TestAccJDCloudAvailabilityGroup_basic(t *testing.T) {

	var agId string
	name1 := randomStringWithLength(10)
	name2 := randomStringWithLength(10)
	description1 := randomStringWithLength(20)
	description2 := randomStringWithLength(20)

	resource.Test(t, resource.TestCase{

		IDRefreshName: "jdcloud_availability_group.terraform_ag",
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccIfAgDestroyed(&agId),
		Steps: []resource.TestStep{
			{
				Config: agConfigSingleAz(name1, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccIfAgExists("jdcloud_availability_group.terraform_ag", &agId),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "availability_group_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "az.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "instance_template_id", "it-fpxvfdch26"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "description", description1),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "ag_type", "kvm"),
				),
			},
			{
				Config: agConfigSingleAz(name2, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccIfAgExists("jdcloud_availability_group.terraform_ag", &agId),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "availability_group_name", name2),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "az.#", "1"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "instance_template_id", "it-fpxvfdch26"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "description", description2),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag", "ag_type", "kvm"),
				),
			},
		},
	})
}

// dual az test
func TestAccJDCloudAvailabilityGroup_dual_az(t *testing.T) {

	var agId string
	name1 := randomStringWithLength(10)
	name2 := randomStringWithLength(10)

	resource.Test(t, resource.TestCase{

		IDRefreshName: "jdcloud_availability_group.terraform_ag_daz",
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccIfAgDestroyed(&agId),
		Steps: []resource.TestStep{
			{
				Config: agConfigDualAz(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccIfAgExists("jdcloud_availability_group.terraform_ag_daz", &agId),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "availability_group_name", name1),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "az.#", "2"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "instance_template_id", "it-fpxvfdch26"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "ag_type", "docker"),

					// Description is set during resource_XXX_read, expected to be "nil"
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "description", ""),
				),
			},
			{
				Config: agConfigDualAz(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccIfAgExists("jdcloud_availability_group.terraform_ag_daz", &agId),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "availability_group_name", name2),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "az.#", "2"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "instance_template_id", "it-fpxvfdch26"),
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "ag_type", "docker"),
						
					// Description is set during resource_XXX_read, expected to be "nil"
					resource.TestCheckResourceAttr(
						"jdcloud_availability_group.terraform_ag_daz", "description", ""),

				),
			},
		},
	})
}

func testAccIfAgExists(agName string, agId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		localAgInfo, ok := stateInfo.RootModule().Resources[agName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfAgExists failed, we can not find a ag namely:{%s} in terraform.State", agName)
		}
		if localAgInfo.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfAgExists failed,operation failed, ag is created but ID not set")
		}

		*agId = localAgInfo.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewAgClient(config.Credential)

		req := apis.NewDescribeAgRequest(config.Region, *agId)
		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.DescribeAg(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {

				if resp.Result.Ag.Name != localAgInfo.Primary.Attributes["availability_group_name"] {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.AgName(%s) != remote AgName(%s)", localAgInfo.Primary.Attributes["availability_group_name"], resp.Result.Ag.Name))
				}
				if resp.Result.Ag.InstanceTemplateId != localAgInfo.Primary.Attributes["instance_template_id"] {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.instance_template_id(%s) != remote instance_template_id(%s)", localAgInfo.Primary.Attributes["instance_template_id"], resp.Result.Ag.InstanceTemplateId))
				}
				if resp.Result.Ag.AgType != localAgInfo.Primary.Attributes["ag_type"] {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.ag_type(%s) != remote ag_type(%s)", localAgInfo.Primary.Attributes["ag_type"], resp.Result.Ag.AgType))
				}
				localAzLength, _ := strconv.Atoi(localAgInfo.Primary.Attributes["az.#"])
				if len(resp.Result.Ag.Azs) != localAzLength {
					resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgExists failed, local.az(%s) != remote az(%s)", localAgInfo.Primary.Attributes["az.#"], resp.Result.Ag.Azs))
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

func testAccIfAgDestroyed(agId *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {

		if *agId == "" {
			return fmt.Errorf("[ERROR] testAccIfAgDestroyed Failed agId is empty")
		}

		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewAgClient(config.Credential)
		req := apis.NewDescribeAgRequest(config.Region, *agId)

		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.DescribeAg(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				return resource.NonRetryableError(fmt.Errorf("[E] testAccIfAgDestroyed failed, resource still exists"))
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
