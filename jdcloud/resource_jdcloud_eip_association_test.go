package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

const TestAccEIPAssociationConfig = `
resource "jdcloud_eip_association" "eip-association-TEST-1"{
	instance_id = "i-g6xse7qb0z"
	elastic_ip_id = "fip-a5uo5mw5al"
}
`

func TestAccJDCloudEIPAssociation_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDiskEIPAssociationDestroy("jdcloud_eip_association.eip-association-TEST-1"),
		Steps: []resource.TestStep{
			{
				Config: TestAccEIPAssociationConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfEIPAssociationExists("jdcloud_eip_association.eip-association-TEST-1"),
				),
			},
		},
	})
}

//-------------------------- Customized check functions

func testAccIfEIPAssociationExists(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[resourceName]
		if ok == false {
			return fmt.Errorf("we can not find a resouce namely:{%s} in terraform.State", resourceName)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, resource:%s is created but ID not set", resourceName)
		}
		EIPId := infoStoredLocally.Primary.Attributes["elastic_ip_id"]
		instanceId := infoStoredLocally.Primary.Attributes["instance_id"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeElasticIpRequest(config.Region, EIPId)
		resp, err := vmClient.DescribeElasticIp(req)

		if err != nil {
			return err
		}

		if resp.Error.Code != REQUEST_COMPLETED || resp.Result.ElasticIp.InstanceId != instanceId {
			return fmt.Errorf("cannot create certain resource")
		}

		return nil
	}
}

func testAccDiskEIPAssociationDestroy(resourceName string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, _ := stateInfo.RootModule().Resources[resourceName]
		EIPId := infoStoredLocally.Primary.Attributes["elastic_ip_id"]
		instanceId := infoStoredLocally.Primary.Attributes["instance_id"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		vmClient := client.NewVpcClient(config.Credential)

		req := apis.NewDescribeElasticIpRequest(config.Region, EIPId)
		resp, err := vmClient.DescribeElasticIp(req)

		if err != nil {
			return err
		}

		if resp.Result.ElasticIp.InstanceId == instanceId {
			return fmt.Errorf("failed in deleting certain resources ")
		}

		return nil
	}
}
