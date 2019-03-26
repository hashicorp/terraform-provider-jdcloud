---
layout: "jdcloud"
page_title: "JDCloud Route Table"
sidebar_current: "docs-jdcloud-resource-route-table"
description: |-
  Provides a JDCloud route table
---
# jdcloud\_route\_table

Provides a JDCloud route table

### Example Usage

```hcl
resource "jdcloud_route_table" "example" {
  vpc_id = "vpc-example"
  route_table_name = "example"
}
```

### Argument Reference

The following arguments are supported:

* `vpc_id` - \(Required\)  :  Route table is supposed to exists under a VPC. Fill the id of this VPC here
* `route_table_name` - \(Required\) : Name this route table 
* `description` - \(Optional\) : Describe it

### Attribute Reference

The following attributes are exported:

* `id` : The id of this route table , used to attach/detach certain route table rules

### Import

Existing route table can be imported to Terraform state by specifying the route table id:

```bash
terraform import jdcloud_route_table.example rtb-abc12345678
```


