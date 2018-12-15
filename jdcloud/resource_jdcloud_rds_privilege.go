package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/models"
)

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
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     privilegeSchema,
			},
		},
	}
}

func resourceJDCloudRDSPrivilegeCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	instanceId := d.Get("instance_id").(string)
	userName := d.Get("username").(string)
	accountPrivileges := []models.AccountPrivilege{}
	for _, item := range d.Get("account_privilege").([]interface{}) {
		itemMap := item.(map[string]interface{})
		accountPrivileges = append(accountPrivileges, models.AccountPrivilege{
			DbName:    stringAddr(itemMap["db_name"]),
			Privilege: stringAddr(itemMap["privilege"]),
		})
	}

	req := apis.NewGrantPrivilegeRequest(config.Region, instanceId, userName, accountPrivileges)
	resp, err := rdsClient.GrantPrivilege(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSPrivilegeCreate failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSPrivilegeCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(userName)
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

			if err := d.Set("account_privilege", latestPrivileges);err!=nil{
				return fmt.Errorf("[ERROR] Failed in resourceJDCloudRDSPrivilegeRead,reasons:%s",err.Error())
			}
			return nil

		}
	}

	d.SetId("")
	return nil
}

func resourceJDCloudRDSPrivilegeDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	instanceId := d.Get("instance_id").(string)
	userName := d.Get("username").(string)
	dbNames := []string{}
	for _, item := range d.Get("account_privilege").([]interface{}) {
		itemMap := item.(map[string]interface{})
		dbNames = append(dbNames, itemMap["db_name"].(string))
	}

	req := apis.NewRevokePrivilegeRequest(config.Region, instanceId, userName, dbNames)
	resp, err := rdsClient.RevokePrivilege(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSPrivilegeDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSPrivilegeDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}
