package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

const vpcTemplate = `
resource "jdcloud_vpc" "%s" {
  vpc_name = "%s"
  cidr_block = "%s"
  description = "%s"
}
`

func generateVPC(resourceName, vpcName, cidr, description string) string {
	return fmt.Sprintf(vpcTemplate, resourceName, vpcName, cidr, description)
}

func copyVPC() {

	vpcClient := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeVpcsRequest(region)
	resp, _ := vpcClient.DescribeVpcs(req)

	for count, vpc := range resp.Result.Vpcs {

		resourceName := fmt.Sprintf("vpc-%d", count)
		vpcResource := generateVPC(resourceName, vpc.VpcName, vpc.AddressPrefix, vpc.Description)
		tracefile(vpcResource)
		resourceMap[vpc.VpcId] = resourceName
	}
}
