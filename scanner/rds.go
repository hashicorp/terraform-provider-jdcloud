package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/models"
	"time"
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

func generatePrivList(instanceID string ,dbIDList []models.AccountPrivilege) string {

	if len(dbIDList) == 0 {
		return "[]"
	}

	ret := "[\n\t"
	for _, dbID := range dbIDList {
		resource := fmt.Sprintf("{db_name = \"%s\",privilege = \"%s\"},\n\t", *dbID.DbName, *dbID.Privilege)
		ret = ret + resource
	}

	return ret + "]"
}

func getAllInstances() (resp *apis.DescribeInstancesResponse,err error) {

	req := apis.NewDescribeInstancesRequest(region)
	c := client.NewRdsClient(config.Credential)

	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err = c.DescribeInstances(req)
		if err == nil && resp.Error.Code == 0 {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return resp,err
}
func performRDSInstanceCopy(instanceID string) (resp *apis.DescribeInstanceAttributesResponse,err error) {

	req := apis.NewDescribeInstanceAttributesRequest(region, instanceID)
	c := client.NewRdsClient(config.Credential)

	err = resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err = c.DescribeInstanceAttributes(req)
		if err == nil && resp.Error.Code == 0 {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return resp,err
}
func performRDSDBCopy(instanceID string) (resp *apis.DescribeDatabasesResponse,err error) {

	reqDB := apis.NewDescribeDatabasesRequest(region, instanceID)
	c := client.NewRdsClient(config.Credential)

	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err = c.DescribeDatabases(reqDB)
		if err == nil && resp.Error.Code == 0 {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return resp,err
}
func performRDSAccountCopy(instanceID string) (resp *apis.DescribeAccountsResponse,err error) {

	reqAcc := apis.NewDescribeAccountsRequest(config.Region, instanceID)
	c := client.NewRdsClient(config.Credential)

	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err = c.DescribeAccounts(reqAcc)
		if err == nil && resp.Error.Code == 0 {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return resp,err
}

func copyRDS() {

	// Get-All available instances
	resp, err := getAllInstances()
	if err != nil || resp.Error.Code!= 0 {
		fmt.Println(formatErrorMessage(resp.Error,err))
		return
	}

	for index, vm := range resp.Result.DbInstances {
		// RDS - Instance
		resourceName := fmt.Sprintf("rds-%d", index)
		resp,err := performRDSInstanceCopy(vm.InstanceId)
		if err != nil || resp.Error.Code!= 0 {
			fmt.Println(formatErrorMessage(resp.Error,err))
			return
		}

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
		respDB, err := performRDSDBCopy(vm.InstanceId)
		if err != nil || respDB.Error.Code!= 0 {
			fmt.Println(formatErrorMessage(resp.Error,err))
			return
		}
		for count, d := range respDB.Result.Databases {
			resourceDbName := fmt.Sprintf("db-%d-%d", index, count)
			tracefile(fmt.Sprintf(rdsDBTemplate, resourceDbName, resourceName, d.DbName, d.CharacterSetName))
		}


		// RDS-Account
		respAcc,err := performRDSAccountCopy(vm.InstanceId)
		if err != nil || respAcc.Error.Code!= 0 {
			fmt.Println(formatErrorMessage(resp.Error,err))
			return
		}
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
				generatePrivList(vm.InstanceId,a.AccountPrivileges)))
		}
	}

}
