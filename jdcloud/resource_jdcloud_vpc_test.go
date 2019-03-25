package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

/*
	TestCase : 1-[Pass].common stuff only. Not yet found any tricky point requires extra attention
*/

const TestAccVpcConfig = `
resource "jdcloud_vpc" "vpc-TEST"{
	vpc_name = "DevOps2019"
	cidr_block = "172.16.0.0/19"
	description = "test"
}
`
const TestAccVpcConfigUpdate = `
resource "jdcloud_vpc" "vpc-TEST"{
	vpc_name = "DevOps2019"
	cidr_block = "172.16.0.0/19"
	description = "testtest"
}
`
const TestAccVpcConfigMin = `
resource "jdcloud_vpc" "vpc-TEST"{
	vpc_name = "DevOps2018"
	cidr_block = "172.16.0.0/19"
}
`

func TestAccJDCloudVpc_basic(t *testing.T) {

	var vpcId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccVpcDestroy(&vpcId),
		Steps: []resource.TestStep{
			{
				Config: TestAccVpcConfigMin,
				Check: resource.ComposeTestCheckFunc(

					testAccIfVpcExists("jdcloud_vpc.vpc-TEST", &vpcId),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "vpc_name", "DevOps2018"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "cidr_block", "172.16.0.0/19"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "description", ""),
				),
			},
			{
				Config: TestAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfVpcExists("jdcloud_vpc.vpc-TEST", &vpcId),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "vpc_name", "DevOps2019"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "cidr_block", "172.16.0.0/19"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "description", "test"),
				),
			},
			{
				Config: TestAccVpcConfigUpdate,
				Check: resource.ComposeTestCheckFunc(

					testAccIfVpcExists("jdcloud_vpc.vpc-TEST", &vpcId),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "vpc_name", "DevOps2019"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "cidr_block", "172.16.0.0/19"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST", "description", "testtest"),
				),
			},
			{
				ResourceName:      "jdcloud_vpc.vpc-TEST",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccIfVpcExists(vpcName string, vpcId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		vpcInfoStoredLocally, ok := stateInfo.RootModule().Resources[vpcName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfVpcExists Failed,we can not find a vpc namely:{%s} in terraform.State", vpcName)
		}
		if vpcInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfVpcExists Failed,operation failed, vpc is created but ID not set")
		}
		vpcIdStoredLocally := vpcInfoStoredLocally.Primary.ID

		vpcConfig := testAccProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(vpcConfig.Credential)

		req := apis.NewDescribeVpcRequest(vpcConfig.Region, vpcIdStoredLocally)
		resp, err := vpcClient.DescribeVpc(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfVpcExists Failed,according to the ID stored locally,we cannot find any VPC on your cloud")
		}

		*vpcId = vpcIdStoredLocally
		return nil
	}
}

func testAccVpcDestroy(vpcIdStoredLocally *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *vpcIdStoredLocally == "" {
			return fmt.Errorf("[ERROR] testAccVpcDestroy Failed,vpcID is empty")
		}

		vpcConfig := testAccProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(vpcConfig.Credential)

		req := apis.NewDescribeVpcRequest(vpcConfig.Region, *vpcIdStoredLocally)
		resp, err := vpcClient.DescribeVpc(req)

		if err != nil {
			return err
		}
		if resp.Error.Code == REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccVpcDestroy Failed,resource still exists,check position-4")
		}
		return nil
	}
}
