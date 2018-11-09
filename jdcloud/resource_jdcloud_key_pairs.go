package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"log"
)

func resourceJDCloudKeyPairs() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudKeyPairsCreate,
		Read:   resourceJDCloudKeyPairsRead,
		Update: resourceJDCloudKeyPairsUpdate,
		Delete: resourceJDCloudKeyPairsDelete,

		Schema: map[string]*schema.Schema{
			"key_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceJDCloudKeyPairsCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	keyName := d.Get("key_name").(string)

	vmClient := client.NewVmClient(config.Credential)

	//构造请求
	rq := apis.NewCreateKeypairRequest(config.Region, keyName)

	//发送请求
	resp, err := vmClient.CreateKeypair(rq)

	if err != nil {

		log.Printf("[DEBUG] create key pairs failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] create key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.RequestID)

	return nil
	//return waitForKeyPairs(d, m, VM_RUNNING)
}

func resourceJDCloudKeyPairsRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceJDCloudKeyPairsUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceJDCloudKeyPairsDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	keyName := d.Get("key_name").(string)

	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDeleteKeypairRequest(config.Region, keyName)
	resp, err := vmClient.DeleteKeypair(req)
	if err != nil {
		log.Printf("[DEBUG]  delete key pairs failed %s", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] delete key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	return nil
}
