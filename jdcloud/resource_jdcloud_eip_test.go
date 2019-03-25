package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"strconv"
	"testing"
	"time"
)

/*
	TestCase : 1-[Pass].common stuff only. Not yet found any tricky point requires extra attention
*/

const TestAccEIPConfig = `
resource "jdcloud_eip" "eip-terraform"{
	eip_provider = "bgp" 
	bandwidth_mbps = 10
}
`

func TestAccJDCloudEIP_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccEIPDestroy("jdcloud_eip.eip-terraform"),
		Steps: []resource.TestStep{
			{
				Config: TestAccEIPConfig,
				Check: resource.ComposeTestCheckFunc(

					// Assigned values
					testAccIfEIPExists("jdcloud_eip.eip-terraform"),
					resource.TestCheckResourceAttr(
						"jdcloud_eip.eip-terraform", "eip_provider", "bgp"),
					resource.TestCheckResourceAttr(
						"jdcloud_eip.eip-terraform", "bandwidth_mbps", "10"),

					// After resource_XYZ_Read these value will be set.
					resource.TestCheckResourceAttrSet(
						"jdcloud_eip.eip-terraform", "elastic_ip_address"),
				),
			},
			{
				ResourceName:      "jdcloud_eip.eip-terraform",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccIfEIPExists(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfEIPExists Failed,we can not find a resouce namely:{%s} in terraform.State", resourceName)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfEIPExists Failed,operation failed, resource:%s is created but ID not set", resourceName)
		}
		eipId := infoStoredLocally.Primary.ID
		resourceId := infoStoredLocally.Primary.Attributes["eip_provider"]
		bandWidth := infoStoredLocally.Primary.Attributes["bandwidth_mbps"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeElasticIpRequest(config.Region, eipId)
		resp, err := vpcClient.DescribeElasticIp(req)

		if err != nil || resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfEIPExists Failed,Error.Code = %d,Error.Message=%s,err.Error()=%s", resp.Error.Code, resp.Error.Message, err.Error())
		}

		bandWidthInt, _ := strconv.Atoi(bandWidth)
		if resp.Result.ElasticIp.Provider != resourceId || resp.Result.ElasticIp.BandwidthMbps != bandWidthInt {
			return fmt.Errorf("[ERROR] testAccIfEIPExists Failed,resource info does not match")
		}

		return nil
	}
}

func testAccEIPDestroy(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, _ := stateInfo.RootModule().Resources[resourceName]
		eipId := infoStoredLocally.Primary.ID

		config := testAccProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeElasticIpRequest(config.Region, eipId)

		for count := 0; count < MAX_EIP_RECONNECT; count++ {

			resp, err := vpcClient.DescribeElasticIp(req)

			if err != nil {
				return fmt.Errorf("[ERROR] testAccEIPDestroy failed %s ", err.Error())
			}

			if resp.Error.Code == RESOURCE_NOT_FOUND {
				return nil
			}
			time.Sleep(3 * time.Second)
		}

		return fmt.Errorf("[ERROR] testAccEIPDestroy failed, resource still exists")
	}
}
