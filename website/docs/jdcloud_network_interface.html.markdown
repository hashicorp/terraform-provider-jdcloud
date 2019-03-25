---
layout: "jdcloud"
page_title: "JDCloud Network Interface"
sidebar_current: "docs-jdcloud-resource-network-interface"
description: |-
  Provides a JDCloud network interface
---

# jdcloud\_network\_interface

Provides a JDCloud network interface

### Example Usage

```hcl
resource "jdcloud_network_interface" "example-ineterface" {
  subnet_id = "subnet-example"
  network_interface_name = "example"
}
```

### Argument Reference

The following arguments are supported:

* `subnet_id` - \(Required\) : Subnet is a string which represents the created network interface belong to
* `network_interface_name` - \(Required\) : The name of this network interface
* `description` - \(Optional\) : Describe this network interface
* `az` - \(Optional\) : The place this network interface locates at
* `primary_ip_address` - \(Optional\) : This is a string of an IP address, specify or leave it blank and a default ip address will be generated and assigned.
* `sanity_check` - \(Optional\) : Idempotent check for this network interface, if you have no idea what this parameter is about, just leave it blank.
* `secondary_ip_addresses` - \(Optional\) : This is a list of private ip addresses,  by the time you create a network interface, you can also specify some private ip addresses with it.
* `secondary_ip_count` - \(Optional\) : If you just want some certain amount of private ip addresses, but not care about what they actually are. Fill in the amount of ips, system will generate them for you.
* `security_groups` - \(Optional\) : A list of security group ids you would like to associate this network interface with.

### Attribute Reference

The following attributes are exported:

* `id` : The id of this network interface, can be used to attach/detach from an instance



