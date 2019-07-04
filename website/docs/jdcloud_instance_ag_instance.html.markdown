---
layout: "jdcloud"
page_title: "JDCloud Ag Instance"
sidebar_current: "docs-jdcloud-ag-instance"
description: |-
Creates instance inside an availability group
---

# jdcloud\_ag\_instance

You can not only creates instance by `jdcloud_instance`, but also through an availability group with a 
instance template. All instance details required are all specified inside this template. All you need 
is just to give it a **unique** name.

### Example Usage

```hcl-terraform
resource "jdcloud_instance_ag_instance" "ag_instance" {
  "availability_group_id" = "ag-example"
  "instances" = [{
    "instance_name" = "xiaohan"
  },{
    "instance_name" = "xiaohan2"
  }]
}
```

### Argument Reference

The following arguments are supported:

* `availability_group_id` - \(Required\): Which availability group you would like to use
* `instances` - \(Required\): A lists consists of the following element
    * `instance_name` - \(Required\): The name of this instance. **Must be unique**

### Attributes Reference

The following attributes are exported:

* `instance_id` - The id of each instance inside this ag, will be used during `ResourceUpdate` and `ResourceDelete`
