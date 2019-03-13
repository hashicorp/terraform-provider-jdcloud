package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

const niTemplate = `
resource "jdcloud_network_interface" "%s"{
	subnet_id = "%s"
	description = "%s"
	az = "%s"
	network_interface_name = "%s"
	secondary_ip_addresses = %s
	security_groups = %s
}
`

func generateList(a []string) string {
	if len(a) == 0 {
		return "[]"
	}
	ret := "["
	for _, item := range a {
		ret = ret + fmt.Sprintf("\"%s\",", item)
	}
	return ret[:len(ret)-1] + "]"
}

func generateReferenceList(a []string, resourceType string) string {
	if len(a) == 0 {
		return "[]"
	}
	ret := "["
	for _, item := range a {
		resource := fmt.Sprintf("\"${%s.%s.id}\",", resourceType, resourceMap[item])
		ret = ret + resource
	}
	return ret[:len(ret)-1] + "]"
}

// NI attachment completed at Instance mode

func copyNetworkInterface() {

	c := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeNetworkInterfacesRequest(region)
	resp, _ := c.DescribeNetworkInterfaces(req)

	for index, ni := range resp.Result.NetworkInterfaces {

		resourceName := fmt.Sprintf("network-interface-%d", index)
		subnetName := fmt.Sprintf("${jdcloud_subnet.%s.id}", resourceMap[ni.SubnetId])

		secondaryIP := []string{}
		for _, ip := range ni.SecondaryIps {
			secondaryIP = append(secondaryIP, ip.PrivateIpAddress)
		}

		if ni.NetworkInterfaceName == "" {
			continue
		}

		tracefile(fmt.Sprintf(niTemplate, resourceName,
			subnetName,
			ni.Description,
			ni.Az,
			ni.NetworkInterfaceName,
			generateList(secondaryIP),
			generateReferenceList(ni.NetworkSecurityGroupIds, "jdcloud_network_interface")))

		resourceMap[ni.NetworkInterfaceId] = resourceName
	}
}
