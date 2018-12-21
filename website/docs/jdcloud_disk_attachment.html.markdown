---
layout: "jdcloud"
page_title: "JDCloud Disk Attachment"
sidebar_current: "docs-jdcloud-disk-attachment"
description: |-
  Helps you to attach an existing disk to an instance
---

# jdcloud\_disk\_attachment

After a disk has been created, you probably need to attach it to an instance

### Example Usage 

```hcl
resource "jdcloud_disk_attachment" "disk-attachment-example"{
  instance_id = "i-example"
  disk_id = "vol-example"
}
```

### Argument Reference

The following arguments are supported:

* `instance_id` - \(Required\): The id of target instance
* `disk_id` - \(Required\): The id of disk
* `auto_delete` - \(Optional\): If this field is set to true, disk will be deleted after it has been detached from  instance
* `device_name` - \(Optional\) : Specify the logical attachment point , for example, attachment point can be "vba" "vbc" etc. Just to make sure this point is available with no other device using it.





