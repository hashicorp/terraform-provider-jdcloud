package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

const eipTemplate = `
resource "jdcloud_eip" "%s"{
  eip_provider = "%s"
  bandwidth_mbps = %d
}
`

// EIP Association completed at Instance part

func copyEIP() {

	c := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeElasticIpsRequest(region)
	resp, _ := c.DescribeElasticIps(req)

	for index, eip := range resp.Result.ElasticIps {
		resourceName := fmt.Sprintf("eip-%d", index)
		tracefile(fmt.Sprintf(eipTemplate, resourceName, eip.Provider, eip.BandwidthMbps))
		resourceMap[eip.ElasticIpId] = resourceName
	}

}
