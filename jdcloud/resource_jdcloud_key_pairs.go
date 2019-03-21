package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	commonModels "github.com/jdcloud-api/jdcloud-sdk-go/services/common/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"io/ioutil"
	"log"
	"os"
	"time"
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
			"key_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudKeyPairsCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	keyName := d.Get("key_name").(string)

	vmClient := client.NewVmClient(config.Credential)

	if publicKey, ok := d.GetOk("public_key"); ok {

		e := resource.Retry(time.Minute, func() *resource.RetryError {
			rq := apis.NewImportKeypairRequest(config.Region, keyName, publicKey.(string))
			resp, err := vmClient.ImportKeypair(rq)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				d.SetId(resp.RequestID)
			}
			if err == nil && resp.Error.Code != REQUEST_COMPLETED {
				return resource.NonRetryableError(fmt.Errorf("[ERROR] import key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message))
			}

			if connectionError(err) {
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(err)
			}
		})
		if e != nil {
			return e
		}

	} else {

		e := resource.Retry(time.Minute, func() *resource.RetryError {
			rq := apis.NewCreateKeypairRequest(config.Region, keyName)
			resp, err := vmClient.CreateKeypair(rq)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {

				// Though setting anything except id is not recommended
				// Private Key can only be retrieved in creating
				d.Set("private_key", resp.Result.PrivateKey)

				if file, ok := d.GetOk("key_file"); ok {

					errIO := ioutil.WriteFile(file.(string), []byte(resp.Result.PrivateKey), KEYPAIRS_PERM)
					if errIO != nil {
						return resource.NonRetryableError(fmt.Errorf("[ERROR] resourceJDCloudKeyPairsCreate failed with error message:%s", errIO.Error()))
					}

					errChmod := os.Chmod(file.(string), KEYPAIRS_PRIV)
					if errChmod != nil {
						return resource.NonRetryableError(fmt.Errorf("[ERROR] resourceJDCloudKeyPairsCreate failed with error message:%s", errChmod.Error()))
					}
				}
				d.SetId(resp.Result.KeyName)
			}
			if err == nil && resp.Error.Code != REQUEST_COMPLETED {
				return resource.NonRetryableError(fmt.Errorf("[DEBUG] create key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message))
			}

			if connectionError(err) {
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(fmt.Errorf("[DEBUG] create key pairs failed %s ", err.Error()))
			}
		})

		if e != nil {
			return e
		}
	}
	return resourceJDCloudKeyPairsRead(d, meta)
}

func resourceJDCloudKeyPairsRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	filters := []commonModels.Filter{
		commonModels.Filter{
			Name:   "keyNames",
			Values: []string{d.Get("key_name").(string)},
		},
	}
	req := apis.NewDescribeKeypairsRequestWithAllParams(config.Region, nil, nil, filters)
	resp, err := vmClient.DescribeKeypairs(req)

	if err != nil {
		return fmt.Errorf("[DEBUG] resourceJDCloudKeyPairsUpdate failed %s", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND || resp.Result.TotalCount == RESOURCE_EMPTY {
		log.Printf("Resource not found, probably have been deleted")
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[DEBUG] resourceJDCloudKeyPairsUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.Set("key_finger_print", resp.Result.Keypairs[0].KeyFingerprint)
	d.Set("key_name", resp.Result.Keypairs[0].KeyName)
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
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[DEBUG] delete key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}
