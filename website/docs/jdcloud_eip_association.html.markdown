---
layout: "jdcloud"
page_title: "JDCloud EIP Association"
sidebar_current: "docs-jdcloud-resource-rip-association"
description: |-
  This helps you associate an EIP with an instance.
---

# jdcloud\_eip\_association

After you have created an EIP, you can attach this EIP with an instance

### Example Usage

```hcl
resource "jdcloud_eip_association" "eip-association-example"{
  instance_id = "i-example"
  elastic_ip_id = "fip-example"
}
```

### Argument Reference 

The following arguments are supported:

* `instance_id` - \(Required\) : The id of instance 
* `elastic_ip_id` - \(Required\): The id of EIP



