package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	disk "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"time"
)

//---------------------------------------------------------------------------------	DISK-SCHEMA-HELPERS

// This function will return the latest status of a disk, level 0
func diskStatusRefreshFunc(d *schema.ResourceData, meta interface{}, diskId string) resource.StateRefreshFunc {

	return func() (diskItem interface{}, diskState string, e error) {

		err := resource.Retry(time.Minute, func() *resource.RetryError {

			config := meta.(*JDCloudConfig)
			c := client.NewDiskClient(config.Credential)
			req := apis.NewDescribeDiskRequest(config.Region, diskId)

			resp, err := c.DescribeDisk(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				diskState = resp.Result.Disk.Status
				diskItem = resp.Result.Disk
				return nil
			}

			if connectionError(err) {
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(err)
			}

		})

		if err != nil {
			return nil, "", err
		}

		return diskItem, diskState, nil
	}
}

// This function will be used here and instance, level 0
// This function does not wait until DISK=Available, it just send request
func performDiskCreate(d *schema.ResourceData, meta interface{}, spec *disk.DiskSpec) (id string, e error) {

	config := meta.(*JDCloudConfig)
	c := client.NewDiskClient(config.Credential)
	req := apis.NewCreateDisksRequest(config.Region, spec, MAX_DISK_COUNT, diskClientTokenDefault())

	e = RetryWithParamsSpecified(2*time.Second, time.Minute, func() *resource.RetryError {

		resp, err := c.CreateDisks(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			id = resp.Result.DiskIds[0]
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	return id, e
}

// This function will send a delete request, level 0
// It does not wait until a disk is completed removed, just send a request
func performDiskDelete(d *schema.ResourceData, meta interface{}, id string) error {

	config := meta.(*JDCloudConfig)
	c := client.NewDiskClient(config.Credential)
	req := apis.NewDeleteDiskRequest(config.Region, id)

	return resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err := c.DeleteDisk(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}
		if connectionError(err) || resp.Error.Code == REQUEST_INVALID {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

}

// This function will wait until a disk is Available, level 1 -> based on diskStatusRefreshFunc
// Disk-Creation usually take couple of minutes, let's wait for it :)
func diskStatusWaiter(d *schema.ResourceData, meta interface{}, id string, pending, target []string) (err error) {

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    diskStatusRefreshFunc(d, meta, id),
		Delay:      3 * time.Second,
		Timeout:    2 * time.Minute,
		MinTimeout: 1 * time.Second,
	}
	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in creatingDisk/Waiting disk,err message:%v", err)
	}
	return nil
}

//----------------------------------------------------------------------------------DISK-SCHEMA-CRUD

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
			"az": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringCandidates("cn-north-1a", "cn-east-1a", "cn-east-1b", "cn-south-1a"),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disk_size_gb": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateDiskSize(),
			},
			"disk_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringCandidates("premium-hdd", "ssd"),
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				// Postpaid_by_usage unavailable in Disk
				ValidateFunc: validateStringCandidates("prepaid_by_duration", "postpaid_by_duration"),
				ForceNew: true,
			},
			"charge_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringCandidates("month", "year"),
				ForceNew:     true,
			},
		},
	}
}

func resourceJDCloudDiskCreate(d *schema.ResourceData, meta interface{}) error {

	diskSpec := disk.DiskSpec{
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

	id, e := performDiskCreate(d, meta, &diskSpec)
	if e != nil {
		return e
	}
	e = diskStatusWaiter(d, meta, id, []string{DISK_CREATING}, []string{DISK_AVAILABLE})
	if e != nil {
		return e
	}
	d.SetId(id)
	// This part is added since attribute "description"
	// Can only be via DiskUpdate rather than "create"
	return resourceJDCloudDiskUpdate(d, meta)
}

func resourceJDCloudDiskRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)
	req := apis.NewDescribeDiskRequestWithAllParams(config.Region, d.Id())

	return resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := diskClient.DescribeDisk(req)

		// Error happens -> finish this round
		if err != nil {
			if connectionError(err) {
				return resource.RetryableError(formatConnectionErrorMessage())
			} else {
				return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
			}
		}

		// Resp.Error non nil -> finish this round
		if resp.Error.Code != REQUEST_COMPLETED {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}

		// All fine -> Disk found deleted -> remove this resource
		if resp.Result.Disk.Status == DISK_DELETED {
			d.SetId("")
			return nil
		}

		// Everything works fine
		d.Set("az", resp.Result.Disk.Az)
		d.Set("name", resp.Result.Disk.Name)
		d.Set("disk_type", resp.Result.Disk.DiskType)
		d.Set("description", resp.Result.Disk.Description)
		d.Set("snapshot_id", resp.Result.Disk.SnapshotId)
		d.Set("charge_mode", resp.Result.Disk.Charge.ChargeMode)
		d.Set("disk_size_gb", resp.Result.Disk.DiskSizeGB)
		return nil
	})
}

func resourceJDCloudDiskUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("name") || d.HasChange("description") {

		config := meta.(*JDCloudConfig)
		diskClient := client.NewDiskClient(config.Credential)
		req := apis.NewModifyDiskAttributeRequestWithAllParams(config.Region, d.Id(), GetStringAddr(d, "name"), GetStringAddr(d, "description"))

		e := resource.Retry(20*time.Second, func() *resource.RetryError {

			resp, err := diskClient.ModifyDiskAttribute(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
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
	}
	return resourceJDCloudDiskRead(d, meta)
}

func resourceJDCloudDiskDelete(d *schema.ResourceData, meta interface{}) error {

	e := performDiskDelete(d, meta, d.Id())
	if e != nil {
		return e
	}

	e = diskStatusWaiter(d, meta, d.Id(), []string{DISK_DELETING}, []string{DISK_DELETED})
	if e != nil {
		return e
	}

	d.SetId("")
	return nil
}
