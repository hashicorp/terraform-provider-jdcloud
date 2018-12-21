---
layout: "jdcloud"
page_title: "JDCloud RDS Database"
sidebar_current: "docs-jdcloud-resource-rds-database"
description: |-
  This helps you create databases in an instance
---

# jdcloud\_rds\_database

After a RDS instance has been created, you probably need to create some databases on this instance

```hcl
resource "jdcloud_rds_database" "db-example"{
  instance_id = "mysql-example"
  db_name = " example"
  character_set = "utf8"
}
```

### Argument Reference 

The following arguments are supported:

* `instance_id`- \(Required\) : This field specifies which RDS instance you are using 
* `db_name`- \(Required\) : Name your database. Each database should have a unique name
* `character_set`- \(Required\) : Candidate character set contains: Utf-8 / SQL\_Latin1\_General\_CP1\_CI\_AS . etc



