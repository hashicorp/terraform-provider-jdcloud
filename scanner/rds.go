package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/models"
)

const rdsTemplate = `
resource "jdcloud_rds_instance" "%s"{
  instance_name = "%s"
  engine = "%s"
  engine_version = "%s"
  instance_class = "%s"
  instance_storage_gb = "%d"
  az = "%s"
  vpc_id = "${jdcloud_vpc.%s.id}"
  subnet_id = "${jdcloud_subnet.%s.id}"
  charge_mode = "postpaid_by_usage"
}
`

const rdsDBTemplate = `
resource "jdcloud_rds_database" "%s"{
  instance_id = "${jdcloud_rds_instance.%s.id}"
  db_name = "%s"
  character_set = "%s"
}
`

const rdsAccTemplate = `
resource "jdcloud_rds_account" "%s"{
  instance_id = "${jdcloud_rds_instance.%s.id}"
  username = "%s"
  password = "DevOps2018"
}
`

const rdsPrivilege = `
resource "jdcloud_rds_privilege" "%s" {
  instance_id = "${jdcloud_rds_instance.%s.id}"
  username = "%s"
  account_privilege = %s
}
`

func generatePrivList(dbIDList []models.AccountPrivilege) string {

	if len(dbIDList) == 0 {
		return "[]"
	}

	ret := "[\n\t"
	for _, dbID := range dbIDList {
		resource := fmt.Sprintf("{db_name = \"%s\",privilege = \"%s\"},\n\t",
			fmt.Sprintf("${jdcloud_rds_database.%s.id}", *dbID.DbName),
			*dbID.Privilege)
		ret = ret + resource
	}

	return ret + "]"
}

func copyRDS() {

	req := apis.NewDescribeInstancesRequest(region)
	c := client.NewRdsClient(config.Credential)
	resp, _ := c.DescribeInstances(req)

	for index, vm := range resp.Result.DbInstances {

		// RDS - Instance
		resourceName := fmt.Sprintf("rds-%d", index)
		req := apis.NewDescribeInstanceAttributesRequest(region, vm.InstanceId)
		resp, _ := c.DescribeInstanceAttributes(req)
		tracefile(fmt.Sprintf(rdsTemplate,
			resourceName,
			resp.Result.DbInstanceAttributes.InstanceName,
			resp.Result.DbInstanceAttributes.Engine,
			resp.Result.DbInstanceAttributes.EngineVersion,
			resp.Result.DbInstanceAttributes.InstanceClass,
			resp.Result.DbInstanceAttributes.InstanceStorageGB,
			resp.Result.DbInstanceAttributes.AzId[0],
			resourceMap[resp.Result.DbInstanceAttributes.VpcId],
			resourceMap[resp.Result.DbInstanceAttributes.SubnetId]))
		resourceMap[vm.InstanceId] = resourceName

		// RDS-Database
		reqDB := apis.NewDescribeDatabasesRequest(region, vm.InstanceId)
		respDB, _ := c.DescribeDatabases(reqDB)
		for count, d := range respDB.Result.Databases {

			resourceDbName := fmt.Sprintf("db-%d-%d", index, count)
			tracefile(fmt.Sprintf(rdsDBTemplate, resourceDbName, resourceName, d.DbName, d.CharacterSetName))
		}

		reqAcc := apis.NewDescribeAccountsRequest(config.Region, vm.InstanceId)
		respAcc, _ := c.DescribeAccounts(reqAcc)
		for count, a := range respAcc.Result.Accounts {

			// RDS-Account
			tracefile(fmt.Sprintf(rdsAccTemplate,
				fmt.Sprintf("rds-acc-%d-%d", index, count),
				resourceName, a.AccountName))

			// RDS - Account Privilege
			tracefile(fmt.Sprintf(rdsPrivilege,
				fmt.Sprintf("rds-priv-%d-%d", index, count),
				resourceName,
				a.AccountName,
				generatePrivList(a.AccountPrivileges)))
		}
	}

}
