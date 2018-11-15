package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	diskModels "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"log"
)

func resourceJDCloudDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudDiskCreate,
		Read:   resourceJDCloudDiskRead,
		Update: resourceJDCloudDiskUpdate,
		Delete: resourceJDCloudDiskDelete,

		Schema: map[string]*schema.Schema{
			"client_token": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"az": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStringNoEmpty,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disk_size_gb": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"multi_attachable": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"charge_duration": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"charge_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringInSlice([]string{"prepaid_by_duration", "postpaid_by_usage", "postpaid_by_duration"}, false),
			},
			"charge_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringInSlice([]string{"month", "year"}, false),
			},
			//应该为set
			"disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceJDCloudDiskCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)

	clientToken := d.Get("client_token").(string)
	maxCount := 1 //olny one disk
	diskSpec := diskModels.DiskSpec{
		Az:         d.Get("az").(string),
		DiskSizeGB: d.Get("disk_size_gb").(int),
		DiskType:   d.Get("disk_type").(string),
		Name:       d.Get("name").(string),
		Charge:     nil,
	}

	req := apis.NewCreateDisksRequest(config.Region, &diskSpec, maxCount, clientToken)

	resp, err := diskClient.CreateDisks(req)
	if err != nil {
		log.Printf("[DEBUG]  resourceJDCloudDiskCreate failed %s", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudDiskCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.Result.DiskIds[0])
	d.Set("disk_id", resp.Result.DiskIds[0])

	return nil
}

func resourceJDCloudDiskRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceJDCloudDiskUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceJDCloudDiskDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)

	diskIDs := d.Get("disk_id").(string)

	req := apis.NewDeleteDiskRequest(config.Region, diskIDs)

	resp, err := diskClient.DeleteDisk(req)
	if err != nil {
		log.Printf("[DEBUG]  resourceJDCloudDiskDelete failed %s", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudDiskDelete  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	return nil
}
