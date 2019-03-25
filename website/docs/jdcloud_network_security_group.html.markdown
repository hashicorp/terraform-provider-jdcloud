---
layout: "jdcloud"
page_title: "JDCloud Security Group"
sidebar_current: "docs-jdcloud-resource-security-group"
description: |-
  Provides a JDCloud network security group
---

# jdcloud\_network\_security\_group

Provides a JDCloud network security group

### Example Usage 

```hcl
resource "jdcloud_network_security_group" "sg-example" {
  network_security_group_name = "sg-example"
  vpc_id = "vpc-example"
}
```

### Argument Reference

The following arguments are supported:

* `network_security_group_name` - \(Required\): Name this security group
* `vpc_id` - \(Required\) : Security group is supposed to exist under a vpc, give the id this vpc
* `description` - \(Optional\): Describe this security group

### Attribute Reference

The following attributes are exported:

* `security_group_id` : A string used to identify this security group, needed when attaching/detaching from a network interface



