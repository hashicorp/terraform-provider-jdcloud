package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"time"
)

const eipTemplate = `
resource "jdcloud_eip" "%s"{
  eip_provider = "%s"
  bandwidth_mbps = %d
}
`

// EIP Association completed at Instance part
func performSingleEIPCopy(req *apis.DescribeElasticIpsRequest) (resp *apis.DescribeElasticIpsResponse, err error) {
	c := client.NewVpcClient(config.Credential)
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err = c.DescribeElasticIps(req)
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
func performEIPCopy() (EIPArray []models.ElasticIp, err error) {

	pageSize := 100
	c := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeElasticIpsRequest(region)

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := c.DescribeElasticIps(req)
		if err == nil && resp.Error.Code == 0 {
			totalCount := resp.Result.TotalCount
			for page := 1; page <= totalCount/100+1; page++ {
				reqPage := apis.NewDescribeElasticIpsRequestWithAllParams(region, &page, &pageSize, nil)
				resp, err = performSingleEIPCopy(reqPage)
				if err != nil {
					return resource.NonRetryableError(err)
				}
				for _, item := range resp.Result.ElasticIps {
					EIPArray = append(EIPArray, item)
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

func copyEIP() {

	EIPArray, err := performEIPCopy()
	if err != nil {
		fmt.Println(err)
		return
	}

	for index, eip := range EIPArray {
		resourceName := fmt.Sprintf("eip-%d", index)
		tracefile(fmt.Sprintf(eipTemplate, resourceName, eip.Provider, eip.BandwidthMbps))
		resourceMap[eip.ElasticIpId] = resourceName
	}

}
