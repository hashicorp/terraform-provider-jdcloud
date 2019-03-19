package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"time"
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

func performSingleSGCopy(req *apis.DescribeNetworkSecurityGroupsRequest) (resp *apis.DescribeNetworkSecurityGroupsResponse, err error) {
	c := client.NewVpcClient(config.Credential)
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err = c.DescribeNetworkSecurityGroups(req)
		if err == nil && resp.Error.Code == 0 {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return
}

func performSGCopy() (sgArray []models.NetworkSecurityGroup, err error) {

	pageSize := 100
	c := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeNetworkSecurityGroupsRequest(region)

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := c.DescribeNetworkSecurityGroups(req)
		if err == nil && resp.Error.Code == 0 {
			totalCount := resp.Result.TotalCount
			for page := 1; page <= totalCount/100+1; page++ {
				reqPage := apis.NewDescribeNetworkSecurityGroupsRequestWithAllParams(region, &page, &pageSize, nil)
				resp, err = performSingleSGCopy(reqPage)
				if err != nil {
					return resource.NonRetryableError(err)
				}
				for _, item := range resp.Result.NetworkSecurityGroups {
					sgArray = append(sgArray, item)
				}
			}
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return
}

func copySecurityGroup() {

	sgArray, err := performSGCopy()
	if err != nil {
		fmt.Println(err)
		return
	}

	for index, sg := range sgArray {

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
