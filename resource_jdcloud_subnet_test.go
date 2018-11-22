package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/pkg/errors"
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

const TestAccSubnetConfig = `
resource "jdcloud_subnet" "subnet-TEST-1"{
	vpc_id = "vpc-npvvk4wr5j"
	cidr_block = "10.0.0.0/16"
	subnet_name = "aa"
	description = "test"
}
`

func TestAccJDCloudSubnet_basic(t *testing.T){

	// This subnet ID is used to create and verify subnet
	// Currently declared but assigned values later
	var subnetId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy(&subnetId),
		Steps: []resource.TestStep{
			{
				Config: TestAccSubnetConfig,
				Check: resource.ComposeTestCheckFunc(

					// SUBNET_ID verification
					testAccIfSubnetExists("jdcloud_subnet.subnet-TEST-1", &subnetId),
					// Remaining attributes validation
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST-1", "vpc_id", "vpc-npvvk4wr5j"),
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST-1", "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST-1", "subnet_name", "aa"),
					resource.TestCheckResourceAttr("jdcloud_subnet.subnet-TEST-1", "description", "test"),
				),
			},
		},
	})

}



func testAccIfSubnetExists(subnetName string,subnetId *string) resource.TestCheckFunc{

	return func(stateInfo *terraform.State) error {

		//STEP-1 : Check if subnet resource has been created locally
		subnetInfoStoredLocally,ok := stateInfo.RootModule().Resources[subnetName]
		if ok==false{
			return fmt.Errorf("subnet namely {%s} has not been created",subnetName)
		}
		if subnetInfoStoredLocally.Primary.ID==""{
			return fmt.Errorf("operation failed, resources created but ID not set")
		}
		subnetIdStoredLocally := subnetInfoStoredLocally.Primary.ID

		//STEP-2 : Check if subnet resource has been created remotely
		subnetConfig := testAccProvider.Meta().(*JDCloudConfig)
		subnetClient := client.NewVpcClient(subnetConfig.Credential)

		req := apis.NewDescribeSubnetRequest(subnetConfig.Region,subnetIdStoredLocally)
		resp, err := subnetClient.DescribeSubnet(req)

		if err!=nil{
			return err
		}
		if resp.Error.Code != 0{
			//return fmt.Errorf("%s",resp.Error)
			return fmt.Errorf("resources created locally but not remotely")
		}

		//  Here subnet resources has been validated to be created locally and
		//  Remotely, next we are going to validate the remaining attributes
		*subnetId = subnetIdStoredLocally
		return nil
 	}
}


func testAccCheckSubnetDestroy(subnetIdStoredLocally *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// subnet ID is not supposed to be empty during testing stage
		if*subnetIdStoredLocally=="" {
			return errors.New("subnetID is empty")
		}

		subnetConfig := testAccProvider.Meta().(*JDCloudConfig)
		subnetClient := client.NewVpcClient(subnetConfig.Credential)

		req := apis.NewDescribeVpcRequest(subnetConfig.Region, *subnetIdStoredLocally)
		resp, err := subnetClient.DescribeVpc(req)

		// ErrorCode is supposed to be 404 since the subnet has already been deleted
		// err is supposed to be nil pointer since query process shall finish
		if err!=nil {
			return err
		}
		if resp.Error.Code!=404{
			return fmt.Errorf("something wrong happens or resource still exists")
		}
		return nil
	}
}
