package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

const sgTemplate = `
resource "jdcloud_network_security_group" "%s" {
  network_security_group_name = "%s"
  vpc_id = "${jdcloud_vpc.%s.id}"
}
`
const sgRuleUpperPart = `
resource "jdcloud_network_security_group_rules" "%s" {
	network_security_group_id = "${jdcloud_network_security_group.%s.id}"
	add_security_group_rules = [`

const sgRuleMidPart = `
      {
      address_prefix = "%s"
      direction = "%d"
      from_port = "%d"
      protocol = "%d"
      to_port = "%d"
      },`
const sgRuleLowerPart = `
  ]
}
`

func copySecurityGroup() {

	c := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeNetworkSecurityGroupsRequest(region)
	resp, _ := c.DescribeNetworkSecurityGroups(req)

	for index, sg := range resp.Result.NetworkSecurityGroups {

		// sg - Copy
		resourceName := fmt.Sprintf("sg-%d", index)
		sgResource := fmt.Sprintf(sgTemplate, resourceName, sg.NetworkSecurityGroupName, resourceMap[sg.VpcId])
		tracefile(sgResource)
		resourceMap[sg.NetworkSecurityGroupId] = resourceName

		// sgRule - Copy
		ruleResourceName := fmt.Sprintf("sg-r-%d", index)
		tracefile(fmt.Sprintf(sgRuleUpperPart, ruleResourceName, resourceName))
		for _, r := range sg.SecurityGroupRules {
			tracefile(fmt.Sprintf(sgRuleMidPart, r.AddressPrefix, r.Direction, r.FromPort, r.Protocol, r.ToPort))
		}
		tracefile(sgRuleLowerPart)
	}
}
