---
layout: "jdcloud"
page_title: "JDCloud OSS"
sidebar_current: "docs-jdcloud-resource-oss-bucket"
description: |-
  Provides a JDCloud OSS bucket
---


# jdcloud\_oss\_bucket

Provides a JDCloud OSS bucket

### Example Usage

```hcl
resource "jdcloud_oss_bucket" "oss-example" {
  bucket_name = "oss-example"
}
```

### Argument Reference 

The following arguments are supported:

* `bucket_name` - \(Required\) : Name this bucket
* `acl` - \(Optional\) : This field specify the privilege on this bucket, candidate options are as following:
  * `private` : Owner has full control to this bucket
  * `public-read`: Owner has full control, other people can read from this but no writing is allowed
  * `public-read-write`: Everyone can read/write from this bucket
