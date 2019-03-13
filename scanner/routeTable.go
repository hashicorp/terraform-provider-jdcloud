package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

const routeTableTemplate = `
resource "jdcloud_route_table" "%s" {
  vpc_id = "${jdcloud_vpc.%s.id}"
  route_table_name = "%s"
  description = "%s"
}
`

const rtRuleTemplate = `
resource "jdcloud_route_table_rule" "%s" {
	route_table_id = "${jdcloud_route_table.%s.id}"
	address_prefix = "%s"
	next_hop_id = "%s"
	next_hop_type = "%s"
	priority = "%d"
}
`

const rtAssociation = `
resource "jdcloud_route_table_association" "%s" {
  subnet_id = "${jdcloud_subnet.%s.id}"
  route_table_id = "${jdcloud_route_table.%s.id}"
}
`

func copyRouteTable() {

	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeRouteTablesRequest(region)
	resp, _ := vpcClient.DescribeRouteTables(req)

	for count, rt := range resp.Result.RouteTables {

		// RouteTable - Copying
		resourceName := fmt.Sprintf("route-table-%d", count)
		rtResource := fmt.Sprintf(routeTableTemplate, resourceName, resourceMap[rt.VpcId], rt.RouteTableName, rt.Description)
		tracefile(rtResource)
		resourceMap[rt.RouteTableId] = resourceName

		// RouteTableRules - Copying
		for ruleIndex, rule := range rt.RouteTableRules {

			ruleName := fmt.Sprintf("routetablerule-%d", ruleIndex)
			ruleResource := fmt.Sprintf(rtRuleTemplate, ruleName, resourceName, rule.AddressPrefix, rule.NextHopId, rule.NextHopType, rule.Priority)
			tracefile(ruleResource)
			resourceMap[rule.RuleId] = ruleName
		}

		// RouteTableAssociation - Copying
		for asIndex, as := range rt.SubnetIds {
			associationName := fmt.Sprintf("route-table-association-%d", asIndex)
			asscoResource := fmt.Sprintf(rtAssociation, associationName, resourceMap[as], resourceName)
			tracefile(asscoResource)
		}
	}
}
