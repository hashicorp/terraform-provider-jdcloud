---
layout: "jdcloud"
page_title: "JDCloud SG Rules"
sidebar_current: "docs-jdcloud-resource-security-group-rules"
description: |-
  This helps to create security group rules
---

# jdcloud\_network\_security\_group\_rules

Provides a JDCloud network security group rules. After you have created security group, this resource helps to assign certain rules with it.

### Example Usage

The following arguments are supported:

```hcl
resource "jdcloud_network_security_group_rules" "sg-rule-example" {
  security_group_id = "sg-example"
  security_group_rules = [{
    address_prefix = "0.0.0.0/0"
    direction      = "0"
    from_port      = "10"
    protocol       = "6"
    to_port        = "20"
  }]
}
```

### Argument Reference

The following arguments are supported:

* `security_group_id` - \(Required\) : The security group that these rules associate with
* `security_group_rules` - \(Required\) :  A list of rule specs where rule spec are defined as following
  * `address_prefix` - \(Required\) : Address prefix look like `10.0.0.0/16` ,which represents , IPs with such form will apply this rule. So this is also a filter for the oncoming IPs.
  Especially, `0.0.0.0/0`means all IPs apply this rule.
  * `direction` - \(Required\) : Data can come into VPC , or get out of VPC, use "direction" to indicate which direction of data apply this rule, candidate value contains:
   <u>**0**</u>-represents data coming in, <u>**1**</u>-represents data going out
  * `protocol` - \(Required\) : Which type of protocol apply this rule, candidate protocol contains, this value can be, candidate value contains:
   <u>**300**</u>-represents all protocol apply this rule, <u>**6**</u>-TCP apply this rule, <u>**17**</u>-UDP apply this rule, <u>**1**</u>-ICMP apply this rule.
  * `from_port` - \(Optional\) : Restriction on starting port, if protocol is transport layer protocol , default value for this field is 1. If not a transport layer protocol , this field will be fixed to 0
  * `to_port` - \(Optional\) : Restriction on ending port, if protocol is transport layer protocol , default value for this field is 65535. If not a transport layer protocol , this field will be fixed to 0
  * `description` - \(Optional\) :  Describe this security group rule

### Attribute Reference

The following attributes are exported:

* `rule_id` : Each rule has its own id for attaching/detaching purpose.

### Import

Existing security group rules can be imported to Terraform state by specifying the id of security group.

```text
terraform import jdcloud_network_security_group_rules.example sg-abc123456
```



