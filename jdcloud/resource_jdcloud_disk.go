package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	diskModels "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
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

	req := apis.NewCreateDisksRequest(config.Region, &diskSpec, MAX_DISK_COUNT, diskClientTokenDefault())
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

	// Disk-Creation usually take couple of minutes, let's wait for it :)
	reqRefresh := apis.NewDescribeDiskRequestWithAllParams(config.Region, d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:    []string{DISK_CREATING},
		Target:     []string{DISK_AVAILABLE},
		Refresh:    diskStatusRefreshFunc(reqRefresh, diskClient),
		Timeout:    3 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in creatingDisk/Waiting disk,err message:%v", err)
	}

	// This part is added since attribute "description"
	// Can only be via DiskUpdate rather than "create"
	return resourceJDCloudDiskUpdate(d, meta)
}

func resourceJDCloudDiskRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)
	req := apis.NewDescribeDiskRequestWithAllParams(config.Region, d.Id())

	err := resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err := diskClient.DescribeDisk(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {

			d.Set("az", resp.Result.Disk.Az)
			d.Set("name", resp.Result.Disk.Name)
			d.Set("disk_type", resp.Result.Disk.DiskType)
			d.Set("description", resp.Result.Disk.Description)
			d.Set("snapshot_id", resp.Result.Disk.SnapshotId)
			d.Set("charge_mode", resp.Result.Disk.Charge.ChargeMode)
			d.Set("disk_size_gb", resp.Result.Disk.DiskSizeGB)
			return nil
		}

		if resp.Result.Disk.Status == DISK_DELETED {
			d.SetId("")
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

	return nil
}

func resourceJDCloudDiskUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("name") || d.HasChange("description") {

		config := meta.(*JDCloudConfig)
		diskClient := client.NewDiskClient(config.Credential)
		req := apis.NewModifyDiskAttributeRequestWithAllParams(config.Region, d.Id(), GetStringAddr(d, "name"), GetStringAddr(d, "description"))

		return resource.Retry(20*time.Second, func() *resource.RetryError {

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
	}
	return resourceJDCloudDiskRead(d, meta)
}

func resourceJDCloudDiskDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	diskClient := client.NewDiskClient(config.Credential)
	req := apis.NewDeleteDiskRequest(config.Region, d.Id())

	err := resource.Retry(time.Minute, func() *resource.RetryError {
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

	reqRefresh := apis.NewDescribeDiskRequestWithAllParams(config.Region, d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:    []string{DISK_DELETING},
		Target:     []string{DISK_DELETED},
		Refresh:    diskStatusRefreshFunc(reqRefresh, diskClient),
		Timeout:    3 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in deletingDisk/Waiting disk,err message:%v", err)
	}

	d.SetId("")
	return nil
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

func diskStatusRefreshFunc(req *apis.DescribeDiskRequest, c *client.DiskClient) resource.StateRefreshFunc {
	return func() (diskItem interface{}, diskState string, e error) {

		err := resource.Retry(3*time.Minute, func() *resource.RetryError {
			resp, err := c.DescribeDisk(req)
			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				diskState = resp.Result.Disk.Status
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

		return nil, diskState, nil
	}
}
