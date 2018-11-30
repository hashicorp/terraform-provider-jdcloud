package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
)

func resourceJDCloudKeyPairs() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudKeyPairsCreate,
		Read:   resourceJDCloudKeyPairsRead,
		Delete: resourceJDCloudKeyPairsDelete,

		Schema: map[string]*schema.Schema{
			"key_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"key_finger_print": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_key": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceJDCloudKeyPairsCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	keyName := d.Get("key_name").(string)

	vmClient := client.NewVmClient(config.Credential)

	if publicKey, ok := d.GetOk("public_key"); ok {

		rq := apis.NewImportKeypairRequest(config.Region, keyName, publicKey.(string))

		resp, err := vmClient.ImportKeypair(rq)

		if err != nil {
			return fmt.Errorf("[ERROR] import key pairs failed %s ", err.Error())
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] import key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		d.SetId(resp.RequestID)

	} else {

		rq := apis.NewCreateKeypairRequest(config.Region, keyName)
		resp, err := vmClient.CreateKeypair(rq)

		if err != nil {
			return fmt.Errorf("[DEBUG] create key pairs failed %s ", err.Error())
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("[DEBUG] create key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		d.SetId(resp.RequestID)
		d.Set("key_finger_print", resp.Result.KeyFingerprint)
		d.Set("private_key", resp.Result.PrivateKey)

	}

	return nil
}

func resourceJDCloudKeyPairsRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	keyName := d.Get("key_name").(string)
	req := apis.NewDescribeKeypairsRequest(config.Region)

	vmClient := client.NewVmClient(config.Credential)
	resp, err := vmClient.DescribeKeypairs(req)

	if err != nil {
		return nil
	}

	for _, key := range resp.Result.Keypairs {
		if key.KeyName == keyName {
			return nil
		}
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] read key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.SetId("")
	return nil
}

func resourceJDCloudKeyPairsDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	keyName := d.Get("key_name").(string)

	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDeleteKeypairRequest(config.Region, keyName)
	resp, err := vmClient.DeleteKeypair(req)

	if err != nil {
		return fmt.Errorf("[DEBUG]  delete key pairs failed %s", err.Error())
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[DEBUG] delete key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}
