package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	diskModels "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"time"
)

// Only one disk is allowed in a disk resource
const maxDiskCount = 1

// Modification allowed : name,description
// Lead to rebuild : Remaining

func resourceJDCloudDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudDiskCreate,
		Read:   resourceJDCloudDiskRead,
		Update: resourceJDCloudDiskUpdate,
		Delete: resourceJDCloudDiskDelete,

		Schema: map[string]*schema.Schema{
			"client_token": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"az": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStringNoEmpty,
				ForceNew:     true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disk_size_gb": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Required: true,
			},
			"multi_attachable": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"charge_duration": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"charge_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringInSlice([]string{"prepaid_by_duration", "postpaid_by_usage", "postpaid_by_duration"}, false),
				ForceNew:     true,
			},
			"charge_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringInSlice([]string{"month", "year"}, false),
				ForceNew:     true,
			},
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

	var clientToken string
	if clientTokenInterface, ok := d.GetOk("client_token"); ok {
		clientToken = clientTokenInterface.(string)
	} else {
		clientToken = diskClientTokenDefault()
	}

	diskSpec := diskModels.DiskSpec{
		Az:         d.Get("az").(string),
		DiskSizeGB: d.Get("disk_size_gb").(int),
		DiskType:   d.Get("disk_type").(string),
		Name:       d.Get("name").(string),
		Charge:     nil,
	}

	req := apis.NewCreateDisksRequest(config.Region, &diskSpec, maxDiskCount, clientToken)
	resp, err := diskClient.CreateDisks(req)

	if err != nil {
		return fmt.Errorf("[DEBUG]  resourceJDCloudDiskCreate failed %s", err.Error())
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[DEBUG] resourceJDCloudDiskCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.Result.DiskIds[0])
	d.Set("disk_id", resp.Result.DiskIds[0])
	d.Set("client_token", clientToken)

	// This part is added since attribute "description"
	// Can only be via DiskUpdate rather than "create"
	if description, ok := d.GetOk("description"); ok {
		d.Set("description", description.(string))
		return resourceJDCloudDiskUpdate(d, meta)
	}

	return nil
}

func resourceJDCloudDiskRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)
	req := apis.NewDescribeDiskRequestWithAllParams(config.Region, d.Id())
	resp, err := diskClient.DescribeDisk(req)

	if err != nil {
		return err
	}

	if resp.Error.Code == 404 {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] failed in resourceJDCloudDiskRead code:%d message:%s", resp.Error.Code, resp.Error.Message)
	}

	d.Set("name", resp.Result.Disk.Name)
	d.Set("description", resp.Result.Disk.Description)
	return nil
}

func resourceJDCloudDiskUpdate(d *schema.ResourceData, meta interface{}) error {

	//Only NAME and DESCRIPTION is allowed to modify
	if d.HasChange("name") || d.HasChange("description") {

		config := meta.(*JDCloudConfig)
		diskClient := client.NewDiskClient(config.Credential)

		req := apis.NewModifyDiskAttributeRequestWithAllParams(config.Region, d.Id(), GetStringAddr(d, "name"), GetStringAddr(d, "description"))
		resp, err := diskClient.ModifyDiskAttribute(req)

		if err != nil {
			return fmt.Errorf("[ERROR] failed in resourceJDCloudDiskUpdate err:%s", err.Error())
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] failed in resourceJDCloudDiskUpdate code:%d message:%s", resp.Error.Code, resp.Error.Message)
		}
	}
	return nil
}

func resourceJDCloudDiskDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)

	diskIDs := d.Get("disk_id").(string)

	req := apis.NewDeleteDiskRequest(config.Region, diskIDs)

	// cloud disk may take some time to delete hence a retry loop is introduced
	for retryCount := 0; retryCount < 3; retryCount++ {

		resp, err := diskClient.DeleteDisk(req)

		if err == nil && resp.Error.Code == 0 {
			break
		}
		if resp.Error.Message == "Cannot delete disk in status creating" ||
			resp.Error.Message == "Can't delete no charged resource" {
			time.Sleep(3 * time.Second)
			continue
		}
		if resp.Error.Code != 0 || err != nil {
			return fmt.Errorf("[ERROR] failed in resourceJDCloudDiskUpdate code:%d message:%s error:", resp.Error.Code, resp.Error.Message, err.Error())
		}
	}

	return nil
}
