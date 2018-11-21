package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"testing"
)

/*
PROCESS:
	1. [TF_ACC] PreCheck          - If necessary parameters exists
	2. Invoke:resourceVpcCreate   - Create corresponding resource
	3. [TF_ACC] resource.TestStep - Check if resources created and attributes set correctly
	4. Invoke: resourceVpcDelete  - Destroy corresponding resource
	4. [TF_ACC] CheckDestroy      - Check if resources has been destroyed correctly
*/

const TestAccVpcConfig = `
resource "jdcloud_vpc" "vpc-TEST-1"{
	name = "vpc_test"
	cidr_block = "10.0.0.0/19"
	description = "test"
}
`

func TestAccJDCloudVpc_basic(t *testing.T) {

	// vpcId is declared but not assigned any values here
	// It will be assigned value in "acceptanceTestCheckIfVpcExists"
	var vpcId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccVpcDestroy(&vpcId),
		Steps: []resource.TestStep{
			{
				Config: TestAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(

					// VPC_ID validation
					testAccIfVpcExists("jdcloud_vpc.vpc-TEST-1", &vpcId),
					// Remaining attributes validation
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST-1", "name", "vpc_tesing_stage"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST-1", "cidr_block", "10.0.0.0/19"),
					resource.TestCheckResourceAttr("jdcloud_vpc.vpc-TEST-1", "description", "test"),
				),
			},
		},
	})
}

//-------------------------- Customized check functions

// Validate attributes on : VPC_ID
func testAccIfVpcExists(vpcName string, vpcId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// STEP-1 : Check if VPC resource has been created locally
		vpcInfoStoredLocally, ok := stateInfo.RootModule().Resources[vpcName]
		if ok == false {
			return fmt.Errorf("we can not find a vpc namely:{%s} in terraform.State", vpcName)
		}
		if vpcInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed, vpc is created but ID not set")
		}
		vpcIdStoredLocally := vpcInfoStoredLocally.Primary.ID

		// STEP-2 : Check if VPC resource has been created remotely
		vpcConfig := acceptanceTestProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(vpcConfig.Credential)

		req := apis.NewDescribeVpcRequest(vpcConfig.Region, vpcIdStoredLocally)
		resp, err := vpcClient.DescribeVpc(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("according to the ID stored locally,we cannot find any VPC on your cloud")
		}

		// Vpc ID has been validated
		// We are going to validate the remaining attributes - name,cidr,description
		*vpcId = vpcIdStoredLocally
		return nil
	}
}

// Validate if VPC resources has been destroyed correctly
func testAccVpcDestroy(vpcIdStoredLocally *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// If vpcID appears to be empty it seems that
		// Some thing went wrong in the previous step
		if *vpcIdStoredLocally == "" {
			return fmt.Errorf("vpcID is empty")
		}

		vpcConfig := acceptanceTestProvider.Meta().(*JDCloudConfig)
		vpcClient := client.NewVpcClient(vpcConfig.Credential)

		req := apis.NewDescribeVpcRequest(vpcConfig.Region, *vpcIdStoredLocally)
		resp, err := vpcClient.DescribeVpc(req)

		// Error.Code is supposed to be 404 since VPC was actually deleted
		// Meanwhile turns out to be 0, successfully queried. Indicating delete error
		if err != nil {
			return err
		}
		if resp.Error.Code == 0 {
			return errors.New("resource still exists,check position-4")
		}
		return nil
	}
}
