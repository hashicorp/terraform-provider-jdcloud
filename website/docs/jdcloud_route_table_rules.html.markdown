---
layout: "jdcloud"
page_title: "JDCloud Route Table Rules"
sidebar_current: "docs-jdcloud-resource-route-table-rules"
description: |-
  This helps to create rules for a route table
---

# jdcloud\_route\_table\_rules

After a route table is created, you probably need to add some routing rules to this route table

### Example Usage

```hcl
resource "jdcloud_route_table_rules" "rule-example"{
  route_table_id = "rtb-example"
  rule_specs = [

    {
      next_hop_type = "internet"
      next_hop_id   = "internet"
      address_prefix= "10.0.0.0/16"
    },

    {
      next_hop_type = "instance"
      next_hop_id   = "i-example"
      address_prefix= "10.0.2.0/16"
    }]

}
```

### Argument Reference 

The following arguments are supported:

* `route_table_id` - \(Required\): Determine which route table you would like to add rules on
* `rule_specs` - \(Required\): A list of rule specs where each spec is specified as following:
  * `next_hop_type` - \(Required\): After routing, which location you would like to go, candidates contains:
    <u>**instance**</u>-route to an instance,
    <u>**internet**</u>-route to public internet,
    <u>**vpc_peering**</u>-used only in vpc peering connection,
    <u>**bgw**</u>-route to boarder gateway.
  * `next_hop_id` - \(Required\): The id of destination,it can be :
    <u>**instance**</u>-If you would like to a certain instance, leave its instance\_id here,
    <u>**internet**</u>-If you would like to route to internet , fill "internet" here.
  * `address_prefix` - \(Required\): look like `10.0.2.0/16`.Address with such prefix will apply this routing rule
  * `priority` - \(Optional\): Specify of the priority of this rule, value varies from 1 to 255, default value is 100 if you leave it blank

### Attribute Reference

The following attributes are exported:

* `id` : id of each rule can be used to attach/detach from this route table


