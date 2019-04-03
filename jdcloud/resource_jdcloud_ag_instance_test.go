package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	common "github.com/jdcloud-api/jdcloud-sdk-go/services/common/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"testing"
	"time"
)

/*
	TestCase : 1 .common stuff only. Not yet found any tricky point requires extra attention
*/

const testAccAGInstanceTemplate = `
resource "jdcloud_instance_ag_instance" "ag_set" {
  "availability_group_id" = "ag-se7v5jwi7o"
  "instances" = [
    {
      "instance_name" = "ark01"
    },
  ]
}
`

const testAccAGInstanceTemplateUpdate = `
resource "jdcloud_instance_ag_instance" "ag_set" {
  "availability_group_id" = "ag-se7v5jwi7o"
  "instances" = [
    {
      "instance_name" = "ark01"
    },
    {
      "instance_name" = "ark02"
    },
    {
      "instance_name" = "ark03"
    },
  ]
}`

const testAccAGInstanceTemplateUpdate2 = `
resource "jdcloud_instance_ag_instance" "ag_set" {
  "availability_group_id" = "ag-se7v5jwi7o"
  "instances" = [
    {
      "instance_name" = "ark01"
    },
    {
      "instance_name" = "ark03"
    },
  ]
}`

func TestAccJDCloudAGInstance_basic(t *testing.T) {

	var agId string

	resource.Test(t, resource.TestCase{

		IDRefreshName: "jdcloud_instance_ag_instance.ag_set",
		PreCheck:      func() { testAccPreCheck(t) },
		Providers:     testAccProviders,
		CheckDestroy:  testAccIfAgInstanceDestroyed(&agId),
		Steps: []resource.TestStep{
			{
				Config: testAccAGInstanceTemplate,
				Check: resource.ComposeTestCheckFunc(
					testAccIfAgInstanceExists("jdcloud_instance_ag_instance.ag_set", &agId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_ag_instance.ag_set", "availability_group_id", "ag-se7v5jwi7o"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_ag_instance.ag_set", "instances.#", "1"),
				),
			},
			{
				Config: testAccAGInstanceTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccIfAgInstanceExists("jdcloud_instance_ag_instance.ag_set", &agId),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_ag_instance.ag_set", "availability_group_id", "ag-se7v5jwi7o"),
					resource.TestCheckResourceAttr(
						"jdcloud_instance_ag_instance.ag_set", "instances.#", "3"),
				),
			},
		},
	})
}

func testAccIfAgInstanceExists(agName string, agId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		localAgInfo, ok := stateInfo.RootModule().Resources[agName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfAgInstanceExists failed, we can not find a ag namely:{%s} in terraform.State", agName)
		}
		if localAgInfo.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfAgInstanceExists failed,operation failed, ag is created but ID not set")
		}

		*agId = localAgInfo.Primary.ID
		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstancesRequest(config.Region)
		req.Filters = []common.Filter{
			common.Filter{
				Name:   "agId",
				Values: []string{*agId},
			},
		}
		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.DescribeInstances(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
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

func testAccIfAgInstanceDestroyed(agId *string) resource.TestCheckFunc {
	return func(stateInfo *terraform.State) error {

		if *agId == "" {
			return fmt.Errorf("[ERROR] testAccIfAgDestroyed Failed agId is empty")
		}

		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewDescribeInstancesRequest(config.Region)
		req.Filters = []common.Filter{
			common.Filter{
				Name:   "agId",
				Values: []string{*agId},
			},
		}

		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.DescribeInstances(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				if resp.Result.TotalCount > 0 {
					return resource.NonRetryableError(fmt.Errorf("It turns out there are %d instances remaining", resp.Result.TotalCount))
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
