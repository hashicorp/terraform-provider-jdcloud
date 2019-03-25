---
layout: "jdcloud"
page_title: "JDCloud ECS Instance"
sidebar_current: "docs-jdcloud-resource"
description: |-
  Provides a JDCloud ECS instance.
---


# jdcloud\_instance

Provides a JDCloud ECS instance 


~> Currently instance paid by "prepaid\_by\_duration" cannot be deleted before they are expired


### Example Usage 

```hcl
resource "jdcloud_instance" "vm-1" {
  az = "cn-north-1a"
  instance_name = "my-vm-1"
  instance_type = "c.n1.large"
  image_id = "bba85cab-dfdc-4359-9218-7a2de429dd80"
  password = "example_password"
  subnet_id = "example_subnetid"
  network_interface_name = "example_ni_name"
  primary_ip = "172.16.0.13"
  secondary_ips = ["172.16.0.14","172.16.0.15"]
  secondary_ip_count   = 2
  security_group_ids = ["example_SgId"]
  sanity_check = 1

  elastic_ip_bandwidth_mbps = 10
  elastic_ip_provider = "bgp"

  system_disk = {
    disk_category = "local"
    device_name = "vda"
    disk_size_gb =  40
  }

  data_disk = {
    disk_category = "local"
    auto_delete = true
    device_name = "vdb"
  }

  data_disk = {
    disk_category = "cloud"
    device_name = "vdc"
    disk_type = "ssd"
    disk_name = "exampleDisk"
    disk_size_gb = 50
    az = "cn-north-1a"

    auto_delete = true
    disk_name = "vm1-datadisk-1"
    description = "test"
  }
}

```

### Argument Reference

The following arguments are supported:

* `az` - \(Required\) The available zone this ECS instance locates at
* `instance_name` - \(Required\) Instance name is a string consists of no more than 32 characters, available characters contains:
  * Chinese characters
  * alphanumeric characters
  * "\_" and "-" \(Underline and hyphen\)
* `instance_type` - \(Required\) Less than 32 characters, [available instance type](https://docs.jdcloud.com/cn/virtual-machines/instance-type-family)
* `images_id` - \(Required\) Image id used to create this ECS instance, can be public image , private image and cloud market place image.
*  `subnet_id` - \(Required\) The id of a VPC subnet. ECS instance created will be in this VPC 
* `system_disk` - \(Required\) The parameter of your system\_disk contains:

  * `disk_category` - \(Required\): can be local or cloud. Especially when the region of this instance is cn-north-1. Only local disk is available. For other regions, both local and cloud are fine.
  * `disk_size_gb` - \(Required\) : The volume of your disk size, for a local system disk locates at cn-north-1, the volume will be fixed to 40Gb
  * `device_name` - \(Required\) : Specify the logical attachment point , for example, attachment point can be "vba" "vbc" etc. Just to make sure this point is available with no other device using it.

* `data_disk` - \(Optional\) : Similar to system disk. You can also create number of data disks together with your ECS instance. 

  * `disk_category` - \(Required\): A string , can be "local" or "cloud".
  * `device_name` - \(Required\) : Specify the logical attachment point , for example, attachment point can be "vba" "vbc" etc. Just to make sure this point is available with no other device using it.
  * `disk_type` - \(Required\) : Type of this disk,  "ssd" or "prenium-hdd".
  * `disk_size_gb` - \(Required\) : The volume of your disk size, for "ssd", volume varies from 20Gb to 1000 Gb. For "prenium-hdd" disk, volume varies from 20Gb to 3000Gb 
  * `disk_name` - \(Required\): A string used to name this disk
  * `az` - \(Required\): The place this disk will be locate at
  * `auto_delete` - \(Optional\) : Bool value. If this value is set to "true", disk will be deleted when it is detached from instance.
  * `snapshot_id` - \(Optional\) :Fill in if you would like to create this disk from a snapshot.
  * `description` - \(Optional\) : Description of this disk

* `description` - \(Optional\) Description of this ECS instance 
* `password` - \(Optional\) If password of this instance is not set. A default password will be sent to you by email and SMS
* `key_names` - \(Optional\) Name of the key pair used to login to instance. Look like `${jdcloud_key_pairs.key-1.key_name}`
* `primary_ip` - \(Optional\) You can specify an public IP address for this instance. If not specified, default public ip address will be generated and assigned.
* `elastic_ip_bandwidth` - \(Optional\) Specify the bandwidth of your public ip.
* `elastic_ip_provider` - \(Optional\) Name of your ip service provider, can be bgp or no\_bgp, according to the region this instance locates at:
  * cn-north-1 : bgp
  * cn-south-1 : bgp or no\_bgp
  * cn-east-1 : bgp or no\_bgp
  * cn-east-2 : bgp
* `security_group_id` - \(Optional\) A list of security group ids to associate with
* `network_interface_name` - \(Optional\) The id of a network interface, each ECS comes with a elastic network interface. You can leave it as a default name or specify a name you would like to see.
* `sanity_check` - \(Optional\) : Idempotent check for this network interface, if you have no idea what this parameter is about, just leave it blank.
* `secondary_ips` - \(Optional\) A list of private ips. These private ips will be associated with the primary network interface on this instance
* `secondary_ip_count` - \(Optional\) Besides specifying some private ips. By specifying this , a number of private ips will be generated and associated with the network interface.

### Attribute Reference 

The following attributes are exported:

* `instance_id` - The id of this instance, can be used to attach disk, network interface.
* `disk_id` - Ids of data disk, can be used to detach certain cloud disk.




