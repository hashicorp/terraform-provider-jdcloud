package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
)

func resourceJDCloudRDSAccount() *schema.Resource {

	return &schema.Resource{
		Create: resourceJDCloudRDSAccountCreate,
		Read:   resourceJDCloudRDSAccountRead,
		Delete: resourceJDCloudRDSAccountDelete,

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
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudRDSAccountCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewCreateAccountRequest(config.Region, d.Get("instance_id").(string), d.Get("username").(string), d.Get("password").(string))
	resp, err := rdsClient.CreateAccount(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountCreate failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.RequestID)
	return nil
}

func resourceJDCloudRDSAccountRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewDescribeAccountsRequest(config.Region, d.Get("instance_id").(string))
	resp, err := rdsClient.DescribeAccounts(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountRead failed %s ", err.Error())
	}

	if resp.Error.Code == 404 {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	for _, user := range resp.Result.Accounts {
		if user.AccountName == d.Get("username").(string) {
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceJDCloudRDSAccountDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewDeleteAccountRequest(config.Region, d.Get("instance_id").(string), d.Get("username").(string))
	resp, err := rdsClient.DeleteAccount(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountDelete failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}
