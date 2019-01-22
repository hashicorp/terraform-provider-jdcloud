package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	diskModels "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"log"
	"time"
)

// Modification allowed : name,description
// Lead to rebuild : Remaining

func resourceJDCloudDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudDiskCreate,
		Read:   resourceJDCloudDiskRead,
		Update: resourceJDCloudDiskUpdate,
		Delete: resourceJDCloudDiskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"client_token": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
		},
	}
}

func resourceJDCloudDiskCreate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)

	diskSpec := diskModels.DiskSpec{
		Az:         d.Get("az").(string),
		DiskSizeGB: d.Get("disk_size_gb").(int),
		DiskType:   d.Get("disk_type").(string),
		Name:       d.Get("name").(string),
	}
	if _, ok := d.GetOk("snapshot_id"); ok {
		diskSpec.SnapshotId = GetStringAddr(d, "snapshot_id")
	}

	chargeSpec := models.ChargeSpec{}
	if chargeModeInterface, ok := d.GetOk("charge_mode"); ok {

		chargeModeString := chargeModeInterface.(string)
		chargeSpec.ChargeMode = &chargeModeString

		if chargeModeString == "prepaid_by_duration" {

			if _, ok := d.GetOk("charge_unit"); ok {
				chargeSpec.ChargeUnit = GetStringAddr(d, "charge_unit")
			} else {
				return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskCreate, charge_unit invalid")
			}

			if _, ok := d.GetOk("charge_duration"); ok {
				chargeSpec.ChargeUnit = GetStringAddr(d, "charge_duration")
			} else {
				return fmt.Errorf("[ERROR] Failed in resourceJDCloudDiskCreate, charge_duration invalid")
			}
		}

		diskSpec.Charge = &chargeSpec
	}

	var clientToken string
	if clientTokenInterface, ok := d.GetOk("client_token"); ok {
		clientToken = clientTokenInterface.(string)
	} else {
		clientToken = diskClientTokenDefault()
		d.Set("client_token", clientToken)
	}

	req := apis.NewCreateDisksRequest(config.Region, &diskSpec, MAX_DISK_COUNT, clientToken)
	err := resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := diskClient.CreateDisks(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.DiskIds[0])
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

	if err != nil {
		return err
	}

	d.SetPartial("az")
	d.SetPartial("name")
	d.SetPartial("disk_type")
	d.SetPartial("client_token")
	d.SetPartial("disk_size_gb")
	d.SetPartial("charge_mode")
	d.SetPartial("charge_unit")
	d.SetPartial("charge_duration")

	// This part is added since attribute "description"
	// Can only be via DiskUpdate rather than "create"
	if description, ok := d.GetOk("description"); ok {
		d.Set("description", description.(string))
		errUpdateDescription := resourceJDCloudDiskUpdate(d, meta)
		if errUpdateDescription != nil {
			log.Println("[WARN] Resource created but seems failed to attach certain")
			log.Println("[WARN] description, refresh it manually later")
		}
	}

	d.SetPartial("description")
	d.Partial(false)
	return waitForDisk(d, meta, d.Id(), DISK_AVAILABLE)
}

func resourceJDCloudDiskRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)
	req := apis.NewDescribeDiskRequestWithAllParams(config.Region, d.Id())
	resp, err := diskClient.DescribeDisk(req)

	if err != nil {
		return err
	}

	if resp.Result.Disk.Status == DISK_DELETED {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] failed in resourceJDCloudDiskRead code:%d message:%s", resp.Error.Code, resp.Error.Message)
	}

	d.Set("name", resp.Result.Disk.Name)
	d.Set("description", resp.Result.Disk.Description)
	return nil
}

func resourceJDCloudDiskUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("name") || d.HasChange("description") {

		d.Partial(true)
		config := meta.(*JDCloudConfig)
		diskClient := client.NewDiskClient(config.Credential)
		req := apis.NewModifyDiskAttributeRequestWithAllParams(config.Region, d.Id(), GetStringAddr(d, "name"), GetStringAddr(d, "description"))

		return resource.Retry(20*time.Second, func() *resource.RetryError {

			resp, err := diskClient.ModifyDiskAttribute(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				d.SetPartial("name")
				d.SetPartial("description")
				d.Partial(false)
				return nil
			}

			if connectionError(err) {
				return resource.RetryableError(formatConnectionErrorMessage())
			} else {
				return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
			}
		})
	}
	return nil
}

func resourceJDCloudDiskDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)
	diskId := d.Id()
	req := apis.NewDeleteDiskRequest(config.Region, diskId)

	err := resource.Retry(3*time.Minute, func() *resource.RetryError {

		resp, err := diskClient.DeleteDisk(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}

		if connectionError(err) || resp.Error.Code == REQUEST_INVALID {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

	if err != nil {
		return err
	}

	reqWatch := apis.NewDescribeDiskRequestWithAllParams(config.Region, d.Id())
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := diskClient.DescribeDisk(reqWatch)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED && resp.Result.Disk.Status == DISK_DELETED {
			d.SetId("")
			return nil
		}

		if connectionError(err) || resp.Result.Disk.Status != DISK_DELETED {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func waitForDisk(d *schema.ResourceData, meta interface{}, id string, expectedStatus string) error {

	currentTime := int(time.Now().Unix())
	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)
	req := apis.NewDescribeDiskRequestWithAllParams(config.Region, id)
	reconnectCount := 0

	for {

		time.Sleep(3 * time.Second)
		resp, err := diskClient.DescribeDisk(req)

		if resp.Result.Disk.Status == expectedStatus {
			return nil
		}

		if int(time.Now().Unix())-currentTime > DISK_TIMEOUT {
			return fmt.Errorf("[ERROR] resourceJDCloudDiskCreate failed, timeout")
		}

		if err != nil {
			if reconnectCount > MAX_RECONNECT_COUNT {
				return fmt.Errorf("[ERROR] resourceJDCloudRDSWait, MAX_RECONNECT_COUNT Exceeded failed %s ", err.Error())
			}
			reconnectCount++
			continue
		} else {
			reconnectCount = 0
		}

	}
}
