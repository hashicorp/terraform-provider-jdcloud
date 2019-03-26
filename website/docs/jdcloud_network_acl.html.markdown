---
layout: "jdcloud"
page_title: "JDCloud Network Acl"
sidebar_current: "docs-jdcloud-network-acl"
description: |-
  Provides a JDCloud Network Acl
---

# jdcloud\_network\_acl

This resource create a network acl

### Example Usage

```hcl-terraform
resource "jdcloud_network_acl" "acl-test" {
  vpc_id = "vpc-exmaple",
  name = "example-name",
  description = "example-descrption-to-this-acl",
}
```

### Argument Reference 

The following arguments are supported

* `vpc_id`  - \(Required\) : Network acls are supposed to exists under a vpc, fill the id of this vpc here
* `name`  - \(Required\) : Name it
* `description`  - \(Optional\) : A string, give it a description if needed

### Attributes Reference

The following attributes are exported:

* `id` - The id of this Network Acl, which can be used to reference this acl. 

### Import

Existing route table can be imported to Terraform state by specifying the network acl id:

```bash
terraform import jdcloud_network_acl.example acl-abc123456
```


