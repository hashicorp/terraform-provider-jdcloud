package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/pkg/errors"
	"strconv"
	"testing"
)

const TestAccNetWorkInterfaceConfig = `
resource "jdcloud_network_interface" "network-interface-TEST-1"{
	subnet_id = "subnet-j8jrei2981"
	description = "test"
	az = "cn-north-1"
	network_interface_name = "test"
	secondary_ip_addresses = ["10.0.1.0","10.0.2.0"]
	secondary_ip_count = "2"
	security_groups = ["sg-yrd5fa7y55"]
}
`

func TestAccJDCloudNetworkInterface_basic(t *testing.T) {

	// This networkInterfaceId is used to create and verify network interface
	// Currently declared but assigned values later
	var networkInterfaceId string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkInterfaceDestroy(&networkInterfaceId),
		Steps: []resource.TestStep{
			{
				Config: TestAccNetWorkInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(

					// networkInterfaceId verification
					testAccIfNetworkInterfaceExists("jdcloud_network_interface.network-interface-TEST-1", &networkInterfaceId),
					// Remaining attributes validation
					resource.TestCheckResourceAttr("jdcloud_network_interface.network-interface-TEST-1", "network_interface_name", "test"),
					resource.TestCheckResourceAttr("jdcloud_network_interface.network-interface-TEST-1", "az", "cn-north-1"),
					resource.TestCheckResourceAttr("jdcloud_network_interface.network-interface-TEST-1", "description", "test"),
				),
			},
		},
	})

}

func testAccIfNetworkInterfaceExists(networkInterfaceName string, networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		//STEP-1 : Check if networkInterface resource has been created locally
		networkInterfaceInfoStoredLocally, ok := stateInfo.RootModule().Resources[networkInterfaceName]
		if ok == false {
			return fmt.Errorf("networkInterfaceName namely {%s} has not been created", networkInterfaceName)
		}
		if networkInterfaceInfoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("operation failed,networkInterfaceName resources created but ID not set")
		}
		networkInterfaceIdStoredLocally := networkInterfaceInfoStoredLocally.Primary.ID

		//STEP-2: Check if networkInterface resource has been created remotely
		networkInterfaceConfig := testAccProvider.Meta().(*JDCloudConfig)
		networkInterfaceClient := client.NewVpcClient(networkInterfaceConfig.Credential)

		requestOnNetworkInterface := apis.NewDescribeNetworkInterfaceRequest(networkInterfaceConfig.Region, networkInterfaceIdStoredLocally)
		responseOnNetworkInterface, err := networkInterfaceClient.DescribeNetworkInterface(requestOnNetworkInterface)

		if err != nil {
			return err
		}
		if responseOnNetworkInterface.Error.Code != 0 {
			return fmt.Errorf("resources created locally but not remotely: code:%d staus:%s message:%s ",
				responseOnNetworkInterface.Error.Code, responseOnNetworkInterface.Error.Status, responseOnNetworkInterface.Error.Message)
		}

		// STEP-3-Verification on SECONDARY-IP-ADDRESSES

		// Build secondary ip list stored locally
		secondaryIpListLengthLocal, _ := strconv.Atoi(networkInterfaceInfoStoredLocally.Primary.Attributes["secondary_ip_addresses.#"])
		secondaryIpListLocal := make([]string, 0, secondaryIpListLengthLocal)
		for i := 0; i < secondaryIpListLengthLocal; i++ {
			ip_index := "secondary_ip_addresses." + strconv.Itoa(i)
			secondaryIpListLocal = append(secondaryIpListLocal, networkInterfaceInfoStoredLocally.Primary.Attributes[ip_index])
		}

		// Build secondary ip list stored remotely
		secondaryIpListLengthRemote := len(responseOnNetworkInterface.Result.NetworkInterface.SecondaryIps)
		secondaryIpListRemote := make([]string, 0, secondaryIpListLengthRemote)
		for i := 0; i < secondaryIpListLengthRemote; i++ {
			secondaryIpListRemote = append(secondaryIpListRemote, responseOnNetworkInterface.Result.NetworkInterface.SecondaryIps[i].PrivateIpAddress)
		}

		// Compare two ip lists, check if they match

		if ok := sliceABelongToB(secondaryIpListLocal, secondaryIpListRemote); ok == false {
			return fmt.Errorf("secondary ip list stored locally and remotely does not match")
		}

		// STEP-4-Verification on SECURITY-GROUPS
		securityGroupLengthLocal, _ := strconv.Atoi(networkInterfaceInfoStoredLocally.Primary.Attributes["security_groups.#"])
		securityGroupLocal := make([]string, 0, securityGroupLengthLocal)
		for i := 0; i < securityGroupLengthLocal; i++ {
			sg_index := "security_groups." + strconv.Itoa(i)
			securityGroupLocal = append(securityGroupLocal, networkInterfaceInfoStoredLocally.Primary.Attributes[sg_index])
		}
		securityGroupRemote := responseOnNetworkInterface.Result.NetworkInterface.NetworkSecurityGroupIds
		if ok := equalSliceString(securityGroupLocal, securityGroupRemote); ok == false {
			return fmt.Errorf("security group list stored locally and remotely does not match")
		}

		//  Here network Interface resources has been validated to be created locally and
		//  Remotely, next we are going to validate the remaining attributes
		*networkInterfaceId = networkInterfaceIdStoredLocally
		return nil
	}
}

func testAccCheckNetworkInterfaceDestroy(networkInterfaceId *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		// networkInterface ID is not supposed to be empty during testing stage
		if *networkInterfaceId == "" {
			return errors.New("networkInterfaceId is empty")
		}

		networkInterfaceConfig := testAccProvider.Meta().(*JDCloudConfig)
		networkInterfaceClient := client.NewVpcClient(networkInterfaceConfig.Credential)

		requestOnNetworkInterface := apis.NewDescribeNetworkInterfaceRequest(networkInterfaceConfig.Region, *networkInterfaceId)
		responseOnNetworkInterface, err := networkInterfaceClient.DescribeNetworkInterface(requestOnNetworkInterface)

		// ErrorCode is supposed to be 404 since the networkInterface has already been deleted
		// err is supposed to be nil pointer since query process shall finish
		if err != nil {
			return err
		}
		if responseOnNetworkInterface.Error.Code != 404 {
			return fmt.Errorf("something wrong happens or resource still exists")
		}
		return nil
	}
}
