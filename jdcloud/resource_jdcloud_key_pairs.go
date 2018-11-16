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
			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
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

			log.Printf("[DEBUG] import key pairs failed %s ", err.Error())
			return err
		}

		if resp.Error.Code != 0 {
			log.Printf("[DEBUG] import key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
			return errors.New(resp.Error.Message)
		}

		d.SetId(resp.RequestID)

	} else {

		rq := apis.NewCreateKeypairRequest(config.Region, keyName)
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
		d.Set("key_finger_print", resp.Result.KeyFingerprint)
		d.Set("private_key", resp.Result.PrivateKey)

	}

	return nil
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
