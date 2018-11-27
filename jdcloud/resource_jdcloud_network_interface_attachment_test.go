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


const TestAccNetworkInterfaceAttachmentConfig = `
resource "jdcloud_network_interface_attachment" "attachment-TEST-1"{
	instance_id = "i-p3yh27xd3s"
	network_interface_id = "port-p49f4wqq8g"
	auto_delete = "true"
}
`

func TestAccJDCloudNetworkInterfaceAttachment_basic(t *testing.T){

	// This networkInterface ID is used to create and verify subnet
	// Currently declared but assigned values later
	var networkInterfaceId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testAccCheckNetworkInterfaceAttachmentDestroy(&networkInterfaceId),
		Steps: []resource.TestStep{
			{
				Config: TestAccNetworkInterfaceAttachmentConfig,
				Check: resource.ComposeTestCheckFunc(

					// INTERFACE_ID verification
					testAccIfNetworkInterfaceAttachmentExists("jdcloud_network_interface_attachment.attachment-TEST-1", &networkInterfaceId),
				),
			},
		},
	})

}



func testAccIfNetworkInterfaceAttachmentExists(attachmentName string,networkInterfaceId *string) resource.TestCheckFunc{

	return func(stateInfo *terraform.State) error {

		//STEP-1 : Check if attachment resource has been created locally
		attachmentInfoStoredLocally,ok := stateInfo.RootModule().Resources[attachmentName]
		if ok==false{
			return fmt.Errorf("attachment namely {%s} has not been created",attachmentName)
		}

		networkInterfaceIdLocal,ok := attachmentInfoStoredLocally.Primary.Attributes["network_interface_id"]
		if attachmentInfoStoredLocally.Primary.ID=="" || ok==false{
			return fmt.Errorf("operation failed, resources created but ID not set")
		}


		//STEP-2 : Check if subnet resource has been created remotely
		attachmentConfig := testAccProvider.Meta().(*JDCloudConfig)
		attachmentClient := client.NewVpcClient(attachmentConfig.Credential)

		req := apis.NewDescribeNetworkInterfaceRequest(attachmentConfig.Region,networkInterfaceIdLocal)
		resp, err := attachmentClient.DescribeNetworkInterface(req)

		if err!=nil{
			return err
		}
		if resp.Error.Code != 0{
			return fmt.Errorf("resources created locally but not remotely")
		}

		instanceIdLocal  := attachmentInfoStoredLocally.Primary.Attributes["instance_id"]
		instanceIdRemote := resp.Result.NetworkInterface.InstanceId

		if instanceIdLocal != instanceIdRemote {
			return fmt.Errorf("resources locally and remotely does not match")
		}

		//  Here subnet resources has been validated to be created locally and
		//  Remotely, next we are going to validate the remaining attributes
		*networkInterfaceId = networkInterfaceIdLocal
		return nil
	}
}


func testAccCheckNetworkInterfaceAttachmentDestroy(networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// networkInterfaceId is not supposed to be empty during testing stage
		if*networkInterfaceId=="" {
			return errors.New("networkInterfaceId is empty")
		}

		attachmentConfig := testAccProvider.Meta().(*JDCloudConfig)
		attachmentClient := client.NewVpcClient(attachmentConfig.Credential)

		//retry_tag:
		req := apis.NewDescribeNetworkInterfaceRequest(attachmentConfig.Region,*networkInterfaceId)
		resp, err := attachmentClient.DescribeNetworkInterface(req)

		// ErrorCode is supposed to be 404 since the subnet has already been deleted
		// err is supposed to be nil pointer since query process shall finish
		if err!=nil {
			return err
		}
		if resp.Error.Code != 0{
			return fmt.Errorf("something wrong happens or resource still exists")
		}
		return nil
	}
}

func grabResource(stateInfo *terraform.State,resourceName string) (*terraform.ResourceState,error) {
	expectedResource,ok   :=stateInfo.RootModule().Resources[resourceName]
	if !ok {
		return nil,errors.New("cannot grab certain resource")
	}
	return expectedResource,nil
}

func grabResourceAttributes(stateInfo *terraform.ResourceState,key string) interface{} {
	infoInterface,ok := stateInfo.Primary.Attributes[key]
	if !ok {
		return errors.New("cannot grab certain attributes")
	}
	return infoInterface
}