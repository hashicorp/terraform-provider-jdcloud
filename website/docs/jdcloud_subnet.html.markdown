---
layout: "jdcloud"
page_title: "JDCloud Subnet"
sidebar_current: "docs-jdcloud-resource-subnet"
description: |-
  Provides a JDCloud subnet
---

# jdcloud\_subnet

Provides a JDCloud subnet

### Example Usage 

```hcl
resource "jdcloud_subnet" "subnet-exmaple"{
	vpc_id = "vpc-example"
	cidr_block = "10.0.128.0/24"
	subnet_name = "example"
}
```

### Argument Reference 

The following arguments are supported:

* `subnet_name` - \(Required\) : Name this subnet. Naming rules are :  Can't be blank and only supports for Chinese, numbers, capital and lowercase letters, English underline “\_” and hyphen “-”. No more than 32 characters
* `vpc_id` - \(Required\) : Subnets are supposed to exists under a vpc , fill the id of vpc here
* `cidr_block` - \(Required\) : The address interval this subnet contains. Especially the cidr\_block of this subnet can not  have overlap with other subnet in this VPC
* `description` - \(Optional\) : Describe this subnet

### Attribute Reference 

The following attributes are exported:

* `id` : id of this subnet, used to reference this subnet




