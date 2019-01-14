package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
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

func copyDisk() {

	c := client.NewDiskClient(config.Credential)
	req := apis.NewDescribeDisksRequest(region)
	resp, _ := c.DescribeDisks(req)

	for index, disk := range resp.Result.Disks {

		resourceName := fmt.Sprintf("disk-%d", index)
		tracefile(fmt.Sprintf(diskTemplate, resourceName,
			disk.Az, disk.Name, disk.Description,
			disk.DiskType, disk.DiskSizeGB))
		resourceMap[disk.DiskId] = resourceName
	}
}
