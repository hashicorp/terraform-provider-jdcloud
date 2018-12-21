---
layout: "jdcloud"
page_title: "JDCloud VPC"
sidebar_current: "docs-jdcloud-resource-vpc"
description: |-
  Provides a JDCloud VPC
---

# jdcloud\_vpc

Provides a JDCloud VPC

### Example Usage

```hcl
resource "jdcloud_vpc" "vpc-example"{
	vpc_name = "example"
	cidr_block = "172.16.0.0/19"
}
```

### Argument reference

The following arguments are supported:

* `vpc_name` - \(Required\) :  The name of this VPC 
* `cidr_block` - \(Required\) :  Each VPC contains an IP addresses interval. For example, a VPC with cidr\_bock look like `172.16.0.0/19` . IPs of subnet within this VPC are subset of `172.16.0.0/19`.
* `description` - \(Optional\) : Describe this VPC

### Attribute Reference 

The following attributes are exported:

* `vpc_id`-  : The id of this vpc, used to specify this VPC

### Import

Existing VPC can be imported to Terraform state by specifying the id of this VPC.

```text
terraform import jdcloud_vpc.example vpc-example
```



