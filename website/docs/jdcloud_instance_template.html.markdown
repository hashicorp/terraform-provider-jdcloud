---
layout: "jdcloud"
page_title: "JDCloud Instance Template"
sidebar_current: "docs-jdcloud-instance-template"
description: |-
  Provides a JDCloud Instance Template
---

# jdcloud\_instance\_template

Instances can be built from images and templates, this resources helps you to create an `instance template`.
`Instance templates` can useful when using `Availability-Group`

### Example Usage

```hcl-terraform
resource "jdcloud_instance_template" "instance_template" {
  "template_name" = "<Name it as you like>"
  "password" = "<Give it a password>"
  "instance_type" = "g.n2.medium"
  "image_id" = "img-example"
  "bandwidth" = 5
  "description" = "GrandJDcloud"
  "ip_service_provider" = "BGP"
  "charge_mode" = "bandwith"
  "subnet_id" = "subnet-exmaple"
  "security_group_ids" = [" sg-example"]
  "system_disk" = {
    "disk_category" = "local"
    "disk_type" = ""
  }
  "data_disks" = {
    disk_category = "cloud"
  }
}
```
### Argument Reference 

The following arguments are supported

* `template_name`  - \(Required\) : A string, name your instance template
* `password`  - \(Optional\) :  String. Once this filed is set. All instance created from this template will use this password 
* `instance_type`  - \(Required\) :  Specs of this Instance. More available instance type lists [Here](https://docs.jdcloud.com/virtual-machines/instance-type-family)
* `image_id`  - \(Required\) :  A string, which image you would like to use, usually [Ubuntu image or Golang images](https://market.jdcloud.com/#/) are good choices
* `ElasticIP` - \(Optional\) : If you would like a public IP, fill in here
    * `ip_service_provider` - \(Optional\): BGP or NonBGP. Principles of them are the same as creating instance, if you are not sure, leave it blank
    * `charge_mode`  - \(Optional\): Candidates are bandwith and flow. By default its `bandwith`
    *  `bandwidth` - \(Required\) : Integer, ranges from 1 to 200
* `subnet_id`  - \(Required\) :  This field determines which `vpc` and `subnet` instances will be
* `security_group_ids`  - \(Required\) : Slices consists of strings. It states the security-groups on this instance
* `description`  - \(Optional\) : Describe it, Just like other resources.
* `system_disk`  - \(Required\) : Parameters for system\_disk contains
  * `disk_category` - \(Required\): can be local or cloud. Especially when the region of this instance is cn-north-1. Only local disk is available. For other regions, both local and cloud are fine.
  * `disk_size_gb` - \(Required\) : The volume of your disk size, for a local system disk locates at cn-north-1, the volume will be fixed to 40Gb
  * `device_name` - \(Required\) : Specify the logical attachment point , for example, attachment point can be "vba" "vbc" etc. Just to make sure this point is available with no other device using it.
* `data_disks`  - \(Optional\) : 
  * `disk_category` - \(Required\): A string , can be "local" or "cloud".
  * `device_name` - \(Required\) : Specify the logical attachment point , for example, attachment point can be "vba" "vbc" etc. Just to make sure this point is available with no other device using it.
  * `disk_type` - \(Required\) : Type of this disk,  "ssd" or "prenium-hdd".
  * `disk_size_gb` - \(Required\) : The volume of your disk size, for "ssd", volume varies from 20Gb to 1000 Gb. For "prenium-hdd" disk, volume varies from 20Gb to 3000Gb 
  * `disk_name` - \(Required\): A string used to name this disk
  * `az` - \(Required\): The place this disk will be locate at
  * `auto_delete` - \(Optional\) : Bool value. If this value is set to "true", disk will be deleted when it is detached from instance.
  * `snapshot_id` - \(Optional\) :Fill in if you would like to create this disk from a snapshot.
  * `description` - \(Optional\) : Description of this disk
  
### Attribute Reference 

The following attributes are exported:

* `id` - The id of this instance template
