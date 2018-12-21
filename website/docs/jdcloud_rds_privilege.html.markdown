---
layout: "jdcloud"
page_title: "JDCloud RDS Privilege"
sidebar_current: "docs-jdcloud-resource-rds-privilege"
description: |-
  Assign privileges to each account 
---
# jdcloud\_rds\_privilege

Assign privileges to each account 

### Example Usage

```hcl
resource "jdcloud_rds_privilege" "pri-example" {
  instance_id = "mysql-example"
  username = "example"
  account_privilege = [
  
    {db_name = "db1",privilege = "rw"},
    {db_name = "db2",privilege = "ro"},
    
  ]
}
```

### Argument Reference

The following arguments are supported:

* `instance_id`- \(Required\) : This field specifies which RDS instance you are using 
* `username`- \(Required\) : The name of account 
* `account_privilege`- \(Required\) : This is a list of rule specs where each rule spec contains:
  * `db_name` - \(Required\) : Specifies the database 
  * `privilege` - \(Required\) : You can choose "rw" - This account is allowed to both read and write from this database. "ro"-This account is only allowed to read from this accoun



