package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"time"
)

const diskTemplate = `
resource "jdcloud_disk" "%s" {
  az           = "%s"
  name         = "%s"
  description  = "%s"
  disk_type    = "%s"
  disk_size_gb = %d
}
`

func performSingleDiskCopy(req *apis.DescribeDisksRequest) (resp *apis.DescribeDisksResponse, err error) {
	c := client.NewDiskClient(config.Credential)
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err = c.DescribeDisks(req)
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
func performDiskCopy() (diskArray []models.Disk, err error) {

	pageSize := 100
	c := client.NewDiskClient(config.Credential)
	req := apis.NewDescribeDisksRequest(region)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := c.DescribeDisks(req)
		if err == nil && resp.Error.Code == 0 {
			totalCount := resp.Result.TotalCount
			for page := 1; page <= totalCount/100+1; page++ {
				reqPage := apis.NewDescribeDisksRequestWithAllParams(region, &page, &pageSize, nil, nil)
				resp, err = performSingleDiskCopy(reqPage)
				if err != nil {
					return resource.NonRetryableError(err)
				}
				for _, item := range resp.Result.Disks {
					diskArray = append(diskArray, item)
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

func copyDisk() {

	diskArray, err := performDiskCopy()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(diskArray)

	for index, disk := range diskArray {

		resourceName := fmt.Sprintf("disk-%d", index)
		tracefile(fmt.Sprintf(diskTemplate, resourceName,
			disk.Az, disk.Name, disk.Description,
			disk.DiskType, disk.DiskSizeGB))
		resourceMap[disk.DiskId] = resourceName
	}
}
