package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
)

const instanceTemplate = `
resource "jdcloud_instance" "%s" {
  az            = "%s"
  instance_name = "%s"
  instance_type = "%s"
  image_id      = "%s"
  password      = "DevOps2018"

  subnet_id              = "${jdcloud_subnet.%s.id}"
  security_group_ids     = %s
  sanity_check           = %d

  system_disk = {
    disk_category = "%s"
    device_name   = "%s"
    disk_type = "%s"
    disk_size_gb =  %d
  }
}
`

const eipAssoTemplate = `
resource "jdcloud_eip_association" "%s"{
	instance_id = "${jdcloud_instance.%s.id}"
	elastic_ip_id = "${jdcloud_eip.%s.id}"
}
`

const niAssoTemplate = `
resource "jdcloud_network_interface_attachment" "%s"{
  instance_id = "${jdcloud_instance.%s.id}"
  network_interface_id = "${jdcloud_network_interface.%s.id}"
}
`

const diskAssoTemplate = `
resource "jdcloud_disk_attachment" "%s"{
  instance_id = "${jdcloud_instance.%s.id}"
  disk_id = "${jdcloud_disk.%s.id}"
}
`

func copyInstance() {

	c := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstancesRequest(region)
	resp, _ := c.DescribeInstances(req)

	for index, vm := range resp.Result.Instances {

		typeDisk := ""
		sgList := []string{}
		ipList := []string{}
		resourceName := fmt.Sprintf("instance-%d", index)

		for _, sg := range vm.PrimaryNetworkInterface.NetworkInterface.SecurityGroups {
			sgList = append(sgList, sg.GroupId)
		}

		if vm.SystemDisk.DiskCategory == "local" {
			typeDisk = "premium-hdd"
		} else {
			typeDisk = vm.SystemDisk.CloudDisk.DiskType
		}

		// Basic information
		tracefile(fmt.Sprintf(instanceTemplate, resourceName,
			vm.Az, vm.InstanceName, vm.InstanceType,
			vm.ImageId, resourceMap[vm.SubnetId],
			generateReferenceList(sgList, "jdcloud_network_security_group"),
			vm.PrimaryNetworkInterface.NetworkInterface.SanityCheck,
			vm.SystemDisk.DiskCategory, vm.SystemDisk.DeviceName, typeDisk,
			vm.SystemDisk.LocalDisk.DiskSizeGB))
		resourceMap[vm.InstanceId] = resourceName

		// EIP
		for _, ip := range vm.PrimaryNetworkInterface.NetworkInterface.SecondaryIps {
			if ip.ElasticIpId != "" {
				ipList = append(ipList, ip.ElasticIpId)
			}
		}
		if vm.PrimaryNetworkInterface.NetworkInterface.PrimaryIp.ElasticIpId != "" {
			ipList = append(ipList, vm.PrimaryNetworkInterface.NetworkInterface.PrimaryIp.ElasticIpId)
		}
		for _, sec := range vm.SecondaryNetworkInterfaces {
			for _, ip := range sec.NetworkInterface.SecondaryIps {
				if ip.ElasticIpId != "" {
					ipList = append(ipList, ip.ElasticIpId)
				}
			}
			if sec.NetworkInterface.PrimaryIp.ElasticIpId != "" {
				ipList = append(ipList, sec.NetworkInterface.PrimaryIp.ElasticIpId)
			}
		}
		for index2, ipId := range ipList {
			eipName := fmt.Sprintf("niAssociation-%d-%d", index, index2)
			tracefile(fmt.Sprintf(eipAssoTemplate,
				eipName,
				resourceName,
				resourceMap[ipId]))
		}

		// NI
		for index2, ni := range vm.SecondaryNetworkInterfaces {
			niName := fmt.Sprintf("niAssociation-%d-%d", index, index2)
			tracefile(fmt.Sprintf(niAssoTemplate, niName, resourceName, resourceMap[ni.NetworkInterface.NetworkInterfaceId]))
		}

		// Disk
		for index2, disk := range vm.DataDisks {
			dName := fmt.Sprintf("disk-Association-%d-%d", index, index2)
			tracefile(fmt.Sprintf(diskAssoTemplate, dName, resourceName, resourceMap[disk.CloudDisk.DiskId]))
		}
	}
}
