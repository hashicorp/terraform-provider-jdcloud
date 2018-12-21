---
layout: "jdcloud"
page_title: "JDCloud RDS Instance"
sidebar_current: "docs-jdcloud-resource-rds-instance"
description: |-
  Provides a JDCloud RDS instance. 
---

# jdcloud\_rds\_instance

Provides a JDCloud RDS instance. 

### Example Usage

```hcl
resource "jdcloud_rds_instance" "rds_example"{
  instance_name = "example"
  engine = "MySQL"
  engine_version = "5.7"
  instance_class = "db.mysql.s1.micro"
  instance_storage_gb = "20"
  az = "cn-north-1a"
  vpc_id = "vpc-example"
  subnet_id = "subnet-example"
  charge_mode = "postpaid_by_usage"
}
```

### Argument Reference

The following arguments are supported:

* `instance_name`- \(Required\) : Name this RDS instance. Restriction on instance\_name lists following
  * Chinese characters and alphanumeric characters
  * "\_" and "-" \(Underline and hyphen\)
  * No less than 2 characters and no more than 32 characters
* `engine`- \(Required\) : Candidate database engine type lists following
  * MySQL
  * Percona
  * MariaDB
  * SQL-Server
* `engine_version`- \(Required\) :  Select engine version for your cloud database
  * MySQL : 5.6 or 5.7
  * Percona : Only 5.7 available
  * MariaDB : Only 10.2 available
  * SQL-Server : 2008R2 / 2012 / 2014 / 2016
* `instance_class`- \(Required\) : Each RDS instance is indeed an ECS instance. So you also have to choose its specification, for example, core numbers and memory
* `instance_storage_gb`- \(Required\) :  Storage size of this RDS instance 
* `az`- \(Required\) : The place that this RDS instance locates at
* `vpc_id`- \(Required\) : Each instance is supposed to exists under a subnet as well as a vpc,  fill in the id of the vpc in this field.
* `subnet_id`- \(Required\) :  Each instance is supposed to exists under a subnet as well as a vpc, fill in the id of subnet in this field.
* `charge_mode`- \(Required\) : Charge mode can be 
  * prepaid\_by\_duration:  This means you would like to pay for a planned term before using this instance. Especially, you can not delete a RDS instance of "prepaid\_by\_duration" type before they expired. Each account can have at most 5 RDS instance
  * postpaid\_by\_duration: This means that you would like to pay for a unplanned term after using this instance
  * postpaid\_by\_usage:  This means you would like to pay after usage according to the instance spec.
* `charge_unit`- \(Optional\) : Used only when charge mode is "prepaid\_by\_duration", can be "month" or "year", by default this value is "month"
* `charge_duration`- \(Optional\) : Used only when charge\_mode is prepaid\_by\_duration, specifies how long you would like to buy. When charge\_duration is "month", charge\_unit varies from 1 to 9, when duration is "year", charge\_unit varies from 1 to 3.

### Attribute Reference 

The following attributes are exported:

* `rds_id`: The id of this RDS instance, can be used to reference this instance.

### Import 

Existing RDS instance can be imported to Terraform state by specifying the rds\_id:

```
terraform import jdcloud_rds_instance.example mysql-abc123456
```



