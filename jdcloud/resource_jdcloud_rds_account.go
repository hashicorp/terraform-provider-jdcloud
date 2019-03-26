package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"time"
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
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceJDCloudRDSAccountCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewCreateAccountRequest(config.Region, d.Get("instance_id").(string), d.Get("username").(string), d.Get("password").(string))

	e := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.CreateAccount(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.RequestID)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

	if e != nil {
		return e
	}
	return resourceJDCloudRDSAccountRead(d, meta)
}

func resourceJDCloudRDSAccountRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewDescribeAccountsRequest(config.Region, d.Get("instance_id").(string))
	resp, err := rdsClient.DescribeAccounts(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountRead failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSAccountRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	for _, user := range resp.Result.Accounts {
		if user.AccountName == d.Get("username").(string) {
			d.Set("username", user.AccountName)
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

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.DeleteAccount(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId("")
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}
