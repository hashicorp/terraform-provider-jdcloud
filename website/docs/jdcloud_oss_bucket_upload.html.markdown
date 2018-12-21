---
layout: "jdcloud"
page_title: "JDCloud OSS Upload"
sidebar_current: "docs-jdcloud-resource-oss-bucket-upload"
description: |-
  This helps you uploading files to OSS bucket 
---



# jdcloud\_oss\_bucket\_upload

After the OSS bucket has been created, you can upload file to this OSS bucket

### Example Usage

```hcl
resource "jdcloud_oss_bucket_upload" "example" {
  bucket_name = "bucket_example"
  file_name = "/path/to/file"
}
```

### Argument Reference

The following arguments are supported:

* `bucket_name` - \(Required\): Bucket name you would like to upload file
* `file_name` - \(Required\) : Location to the file you would like to upload. For example , this field can be `/home/DevOps/HelloWorld.cpp` In order to avoid unable to find your file, absolute path to this file is recommended to use.



