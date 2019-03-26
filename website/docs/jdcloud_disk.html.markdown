---
layout: "jdcloud"
page_title: "JDCloud Disk"
sidebar_current: "docs-jdcloud-disk"
description: |-
  Provides a JDCloud disk 
---


# jdcloud\_disk

Provides a JDCloud disk 


~> Currently disks paid by "prepaid\_by\_duration" cannot be deleted before they are expired


### Example Usage

```hcl
resource "jdcloud_disk" "example" {
  az           = "cn-north-1a"
  name         = "example_disk"
  description  = "test"
  disk_type    = "ssd"
  disk_size_gb = 60
}
```

### Argument Reference

The following arguments are supported:

* `name` - \(Required\): A string to name this cloud disk
* `disk_size_gb` - \(Required\): The volume of this disk. For a "ssd" disk, the volume varies from 20Gb-2000Gb,  for a "prenium-hdd" disk, the volume varies from 20Gb to 3000Gb.
* `disk_type` - \(Required\): Can be "ssd" or "prenium-hdd"
* `az` - \(Required\):  The place that this disk will be locate at.
* `client_token` - \(Optional\):  Idempotent check parameter. If you have no idea about this parameter, just leave it blank and a default one will be generated.
* `multi-attachable` - \(Optional\): Determine if this disk can be attached to several instance at the same time.
* `description` - \(Optional\):  Describe this disk
* `snapshot_id` - \(Optional\): If you would like to create a disk from an existing snapshot, fill in the id here
* `charge_mode` - \(Optional\): Candidate payment method lists as following :
  * "prepaid\_by\_duration" : Pay before using at a planned interval
  * "postpaid\_by\_usage" : Pay after using, price will be determined according to disk specs and time
  * "postpaid\_by\_duration":Pay after using, price will be determined according to disk specs and time
* `charge_duration` - \(Optional\): Used only when charge\_mode is prepaid\_by\_duration, can be "month" ,"year", default : "month" 
* `charge_unit` - \(Optional\): Used only when charge\_mode is prepaid\_by\_duration, specifies how long you would like to buy. When charge\_duration is "month", charge\_unit varies from 1 to 9, when duration is "year", charge\_unit varies from 1 to 3.

### Attributes Reference

The following attributes are exported:

* `id` - The id of this disk, can be used to attach/detach from an instance, look like vol-xxxx

### Import

Existing disk object can be imported to Terraform state by specifying the disk id:

```text
terraform import jdcloud_disk.example vol-abc12345678
```

