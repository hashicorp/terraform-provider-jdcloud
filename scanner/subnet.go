package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

const subnetTemplate = `
resource "jdcloud_subnet" "%s" {
	vpc_id = "${jdcloud_vpc.%s.id}"
	cidr_block = "%s"
	subnet_name = "%s"
	description = "%s"
}
`

func generateSubnet(resourceName, vpcName, cidr, subnetName, description string) string {
	return fmt.Sprintf(subnetTemplate, resourceName, vpcName, cidr, subnetName, description)
}

func copySubnet() {

	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeSubnetsRequest(region)
	resp, _ := vpcClient.DescribeSubnets(req)

	for count, sn := range resp.Result.Subnets {

		resourceName := fmt.Sprintf("subnet-%d", count)
		subnetResource := generateSubnet(resourceName, resourceMap[sn.VpcId], sn.AddressPrefix, sn.SubnetName, sn.Description)
		tracefile(subnetResource)
		resourceMap[sn.SubnetId] = resourceName
	}
}
