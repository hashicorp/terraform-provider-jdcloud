package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

const TestAccSubnetConfig = `
resource "jdcloud_subnet" "subnet-TEST"{
	vpc_id = "vpc-npvvk4wr5j"
	cidr_block = "10.0.128.0/24"
	subnet_name = "DevOps2018"
	description = "test"
}
`

func TestAccJDCloudSubnet_basic(t *testing.T) {

	var subnetId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy(&subnetId),
		Steps: []resource.TestStep{
			{
				Config: TestAccSubnetConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfSubnetExists("jdcloud_subnet.subnet-TEST", &subnetId),
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST", "vpc_id", "vpc-npvvk4wr5j"),
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST", "cidr_block", "10.0.128.0/24"),
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST", "subnet_name", "DevOps2018"),
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST", "description", "test"),
				),
			},
		},
	})

}

func testAccIfSubnetExists(subnetName string, subnetId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		info, ok := stateInfo.RootModule().Resources[subnetName]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfSubnetExists Failed, subnet namely {%s} has not been created", subnetName)
		}
		if info.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfSubnetExists Failed, operation failed, resources created but ID not set")
		}
		*subnetId = info.Primary.ID

		conf := testAccProvider.Meta().(*JDCloudConfig)
		c := client.NewVpcClient(conf.Credential)

		req := apis.NewDescribeSubnetRequest(conf.Region, *subnetId)
		resp, err := c.DescribeSubnet(req)

		if err != nil || resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfSubnetExists Failed in reading, err:%s, resp:%#v", err.Error(), resp.Error)
		}

		return nil
	}
}

func testAccCheckSubnetDestroy(subnetIdStoredLocally *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *subnetIdStoredLocally == "" {
			return fmt.Errorf("[ERROR] testAccCheckSubnetDestroy Failed,subnetID is empty")
		}

		conf := testAccProvider.Meta().(*JDCloudConfig)
		c := client.NewVpcClient(conf.Credential)

		req := apis.NewDescribeVpcRequest(conf.Region, *subnetIdStoredLocally)
		resp, err := c.DescribeVpc(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != RESOURCE_NOT_FOUND {
			return fmt.Errorf("[ERROR] testAccCheckSubnetDestroy Failed,something wrong happens or resource still exists")
		}

		return nil
	}
}
