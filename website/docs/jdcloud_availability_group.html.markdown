---
layout: "jdcloud"
page_title: "JDCloud AG"
sidebar_current: "docs-jdcloud-availability-group"
description: |-
  Provides a JDCloud Availability Group
---

#jdcloud\_availability\_group

Provides a JDCloud Availability Group

### Example Usage

```hcl
resource "jdcloud_availability_group" "ag_01" {
  availability_group_name = "example_ag_name"
  az = ["cn-north-1a","cn-north-1b"]
  instance_template_id = "example_template_id"
  description  = "This is an example description"
  ag_type = "docker"
}
```

### Argument Reference 

The following arguments are supported

* `availability_group_name` - \(Required\) : A string to name this availability  group 
* `az`  - \(Required\) : Az is a slice consists of strings. All azs has to be in same region, e.g. cn-north-1
* `instance_template_id`  - \(Required\) : A string. All instances within this Ag are created under a certain template,specify this template with its id
* `ag_type`   - \(Optional\) : A string. It decides which kind of 'instance' you are expecting, this fields can be `docker` or `kvm`, by default it is set to `kvm`
*  `description`  - \(Optional\) : Describe this Ag, if needed.

### Attributes Reference

The following attributes are exported:

* `id` - The id of this Ag, can be used to reference this availability group. 