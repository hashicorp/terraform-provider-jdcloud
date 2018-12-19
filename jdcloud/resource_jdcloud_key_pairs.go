package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	commonModels "github.com/jdcloud-api/jdcloud-sdk-go/services/common/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"io/ioutil"
	"log"
	"os"
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

		rq := apis.NewImportKeypairRequest(config.Region, keyName, publicKey.(string))

		resp, err := vmClient.ImportKeypair(rq)

		if err != nil {
			return fmt.Errorf("[ERROR] import key pairs failed %s ", err.Error())
		}

		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] import key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		d.SetId(resp.RequestID)

	} else {

		rq := apis.NewCreateKeypairRequest(config.Region, keyName)
		resp, err := vmClient.CreateKeypair(rq)

		if err != nil {
			return fmt.Errorf("[DEBUG] create key pairs failed %s ", err.Error())
		}

		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[DEBUG] create key pairs failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		d.SetId(resp.Result.KeyName)
		d.Set("key_finger_print", resp.Result.KeyFingerprint)
		d.Set("private_key", resp.Result.PrivateKey)

		if file, ok := d.GetOk("key_file"); ok {

			errIO := ioutil.WriteFile(file.(string), []byte(resp.Result.PrivateKey), KEYPAIRS_PERM)
			if errIO != nil {
				return fmt.Errorf("[ERROR] resourceJDCloudKeyPairsCreate failed with error message:%s", errIO.Error())
			}

			errChmod := os.Chmod(file.(string), KEYPAIRS_PRIV)
			if errChmod != nil {
				return fmt.Errorf("[ERROR] resourceJDCloudKeyPairsCreate failed with error message:%s", errChmod.Error())
			}
		}

	}

	return nil
}

func resourceJDCloudKeyPairsRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	keyName := d.Get("key_name").(string)

	vmClient := client.NewVmClient(config.Credential)

	var filter commonModels.Filter
	filter.Name = "keyNames"
	filter.Values = append(filter.Values, keyName)

	var filters []commonModels.Filter

	filters = append(filters, filter)
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
