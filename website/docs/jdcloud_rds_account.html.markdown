---
layout: "jdcloud"
page_title: "JDCloud RDS Account"
sidebar_current: "docs-jdcloud-resource-rds-account"
description: |-
  This helps you create users for an RDS instance
---
# jdcloud\_rds\_account

Set up users for each RDS instance

### Example Usage

```hcl
resource "jdcloud_rds_account" "account_example"{
  instance_id = "mysql-example"
  username = "example"
  password = "example2018"
}
```

### Argument Reference

The following arguments are supported:

* `instance_id`- \(Required\):  Fill the id of instance you would like to set up account on
* `username`- \(Required\): Your username. Naming rules:  Letters both in upper case and lower case and English underline "\_", no more than 16 characters
* `password`- \(Required\):  Password must contain and only supports letters both in upper case and lower case as well as figures, no less than 8 characters and no more than 16 characters



