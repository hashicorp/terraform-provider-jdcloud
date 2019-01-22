package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/models"
	"time"
)

func typeSetToAccountStructList(set *schema.Set) []models.AccountPrivilege {

	a := []models.AccountPrivilege{}
	for _, account := range set.List() {

		m := account.(map[string]interface{})
		a = append(a, models.AccountPrivilege{
			DbName:    getMapStrAddr(m["db_name"].(string)),
			Privilege: getMapStrAddr(m["privilege"].(string)),
		})
	}
	return a
}

func dbNameList(set *schema.Set) []string {

	dbNameList := []string{}

	for _, db := range set.List() {
		m := db.(map[string]interface{})
		dbNameList = append(dbNameList, m["db_name"].(string))
	}
	return dbNameList
}

func performDetachDB(d *schema.ResourceData, m interface{}, list []string) error {

	config := m.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewRevokePrivilegeRequest(config.Region, d.Get("instance_id").(string), d.Get("username").(string), list)
	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.RevokePrivilege(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return nil
}

func performAttachDB(d *schema.ResourceData, m interface{}, attachSet *schema.Set) error {

	config := m.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewGrantPrivilegeRequest(config.Region, d.Get("instance_id").(string), d.Get("username").(string), typeSetToAccountStructList(attachSet))

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.GrantPrivilege(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func resourceJDCloudRDSPrivilege() *schema.Resource {

	privilegeSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"db_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"privilege": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	return &schema.Resource{
		Create: resourceJDCloudRDSPrivilegeCreate,
		Read:   resourceJDCloudRDSPrivilegeRead,
		Update: resourceJDCloudRDSPrivilegeUpdate,
		Delete: resourceJDCloudRDSPrivilegeDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_privilege": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     privilegeSchema,
				MinItems: 1,
			},
		},
	}
}

func resourceJDCloudRDSPrivilegeCreate(d *schema.ResourceData, m interface{}) error {

	if err := performAttachDB(d, m, d.Get("account_privilege").(*schema.Set)); err != nil {
		return err
	}

	d.SetId(d.Get("username").(string))
	return nil
}

func resourceJDCloudRDSPrivilegeRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewDescribeAccountsRequest(config.Region, d.Get("instance_id").(string))
	resp, err := rdsClient.DescribeAccounts(req)
	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSPrivilegeRead failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSPrivilegeRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	for _, user := range resp.Result.Accounts {

		if user.AccountName == d.Get("username").(string) {

			latestPrivileges := make([]map[string]string, 0, len(user.AccountPrivileges))

			for _, privilege := range user.AccountPrivileges {
				latestPrivilege := map[string]string{
					"db_name":   *privilege.DbName,
					"privilege": *privilege.Privilege,
				}
				latestPrivileges = append(latestPrivileges, latestPrivilege)
			}

			if err := d.Set("account_privilege", latestPrivileges); err != nil {
				return fmt.Errorf("[ERROR] Failed in resourceJDCloudRDSPrivilegeRead,reasons:%s", err.Error())
			}
			return nil

		}
	}

	d.SetId("")
	return nil
}

func resourceJDCloudRDSPrivilegeUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	if d.HasChange("account_privilege") {

		pInterface, cInterface := d.GetChange("account_privilege")
		p := pInterface.(*schema.Set)
		c := cInterface.(*schema.Set)
		i := p.Intersection(c)

		if err := performDetachDB(d, m, dbNameList(p.Difference(i))); err != nil && len(dbNameList(p.Difference(i))) != 0 {
			return err
		}
		if err := performAttachDB(d, m, c.Difference(i)); err != nil && len(dbNameList(c.Difference(i))) != 0 {
			return err
		}

		d.SetPartial("account_privilege")
	}
	d.Partial(false)
	return nil
}

func resourceJDCloudRDSPrivilegeDelete(d *schema.ResourceData, m interface{}) error {

	if err := performDetachDB(d, m, dbNameList(d.Get("account_privilege").(*schema.Set))); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
