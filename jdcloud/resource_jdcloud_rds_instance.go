package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	rds "github.com/jdcloud-api/jdcloud-sdk-go/services/rds/models"
	"time"
)

/*
Parameter Description
	Engine& EngineVersion:  docs.jdcloud.com/cn/rds/api/enum-definitions
   InstanceClass&StorageGB : docs.jdcloud.com/cn/rds/api/instance-specifications-mysql
   ChargeSpec : github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models/ChargeSpec.go
*/

func resourceJDCloudRDSInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudRDSInstanceCreate,
		Read:   resourceJDCloudRDSInstanceRead,
		Update: resourceJDCloudRDSInstanceUpdate,
		Delete: resourceJDCloudRDSInstanceDelete,

		Schema: map[string]*schema.Schema{
			"instance_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"engine": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"engine_version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_class": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"instance_storage_gb": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"az": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"internal_domain_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_domain_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_port": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_mode": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"charge_mode": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"charge_unit": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"charge_duration": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudRDSInstanceCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	chargeSpec := &models.ChargeSpec{}
	chargeMode := d.Get("charge_mode").(string)
	if chargeMode == "prepaid_by_duration" {
		if _, ok := d.GetOk("charge_unit"); !ok {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed in chargeMode parameter")
		}
		if _, ok := d.GetOk("charge_duration"); !ok {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed in chargeDuration parameter")
		}
		chargeSpec = &models.ChargeSpec{ChargeMode: &chargeMode, ChargeUnit: GetStringAddr(d, "charge_unit"), ChargeDuration: GetIntAddr(d, "charge_duration")}
	} else {
		chargeSpec = &models.ChargeSpec{ChargeMode: &chargeMode}
	}

	req := apis.NewCreateInstanceRequest(
		config.Region,
		&rds.DBInstanceSpec{
			InstanceName:      GetStringAddr(d, "instance_name"),
			Engine:            d.Get("engine").(string),
			EngineVersion:     d.Get("engine_version").(string),
			InstanceClass:     d.Get("instance_class").(string),
			InstanceStorageGB: d.Get("instance_storage_gb").(int),
			AzId:              []string{d.Get("az").(string)},
			VpcId:             d.Get("vpc_id").(string),
			SubnetId:          d.Get("subnet_id").(string),
			ChargeSpec:        chargeSpec,
		},
	)

	if validateVPC := verifyVPC(d, meta, d.Get("vpc_id").(string), d.Get("subnet_id").(string)); validateVPC != nil {
		return validateVPC
	}

	instanceId := ""
	rdsClient := client.NewRdsClient(config.Credential)

	// Send a request here
	err := resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.CreateInstance(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			instanceId = resp.Result.InstanceId
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

	// Wait until RDS ready
	err = rdsStatusWaiter(d, meta, instanceId, []string{RDS_CREATING, RDS_DELETED}, []string{RDS_READY})
	if err != nil {
		return err
	}

	d.SetId(instanceId)
	// "domain_name" is needed but only available through reading
	return resourceJDCloudRDSInstanceRead(d, meta)
}

func resourceJDCloudRDSInstanceRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	req := apis.NewDescribeInstanceAttributesRequest(config.Region, d.Id())
	rdsClient := client.NewRdsClient(config.Credential)

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.DescribeInstanceAttributes(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.Set("instance_name", resp.Result.DbInstanceAttributes.InstanceName)
			d.Set("instance_class", resp.Result.DbInstanceAttributes.InstanceClass)
			d.Set("instance_storage_gb", resp.Result.DbInstanceAttributes.InstanceStorageGB)
			d.Set("internal_domain_name", resp.Result.DbInstanceAttributes.InternalDomainName)
			d.Set("public_domain_name", resp.Result.DbInstanceAttributes.PublicDomainName)
			d.Set("instance_port", resp.Result.DbInstanceAttributes.InstancePort)
			d.Set("connection_mode", resp.Result.DbInstanceAttributes.ConnectionMode)
			d.Set("engine", resp.Result.DbInstanceAttributes.Engine)
			d.Set("engine_version", resp.Result.DbInstanceAttributes.EngineVersion)
			d.Set("az", resp.Result.DbInstanceAttributes.AzId[0])
			d.Set("vpc_id", resp.Result.DbInstanceAttributes.VpcId)
			d.Set("subnet_id", resp.Result.DbInstanceAttributes.SubnetId)
			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND || resp.Error.Code == REQUEST_INVALID {
			d.SetId("")
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func resourceJDCloudRDSInstanceUpdate(d *schema.ResourceData, meta interface{}) error {

	d.Partial(true)
	defer d.Partial(false)

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	// Currently you can not degrade your configuration, only upgrade them is allowed
	if d.HasChange("instance_class") || d.HasChange("instance_storage_gb") {
		req := apis.NewModifyInstanceSpecRequest(config.Region, d.Id(), d.Get("instance_class").(string), d.Get("instance_storage_gb").(int))
		resp, err := rdsClient.ModifyInstanceSpec(req)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed %s ", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		err = rdsStatusWaiter(d, meta, d.Id(), []string{RDS_UPDATING, RDS_UNCERTAIN}, []string{RDS_READY})
		if err != nil {
			return err
		}

		d.SetPartial("instance_class")
		d.SetPartial("instance_storage_gb")
	}

	return resourceJDCloudRDSInstanceRead(d, meta)
}

func resourceJDCloudRDSInstanceDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewDeleteInstanceRequest(config.Region, d.Id())

	// Send an DELETE request
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.DeleteInstance(req)

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
	// Wait until this Instance has been completely deleted
	return rdsStatusWaiter(d, meta, d.Id(), []string{RDS_DELETING, RDS_READY}, []string{RDS_DELETED, RDS_UNCERTAIN})
}

func rdsInstanceStatusRefreshFunc(d *schema.ResourceData, meta interface{}, rdsId string) resource.StateRefreshFunc {

	return func() (rds interface{}, rdsState string, e error) {

		err := resource.Retry(time.Minute, func() *resource.RetryError {

			config := meta.(*JDCloudConfig)
			req := apis.NewDescribeInstanceAttributesRequest(config.Region, rdsId)
			rdsClient := client.NewRdsClient(config.Credential)
			rdsClient.SetLogger(vmLogger{})
			resp, err := rdsClient.DescribeInstanceAttributes(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				rds = resp.Result.DbInstanceAttributes
				rdsState = resp.Result.DbInstanceAttributes.InstanceStatus
				return nil
			}

			// 400, In deleting , means rds has already been deleted
			if err == nil && resp.Error.Code == REQUEST_INVALID {
				rds = "RDS"
				rdsState = RDS_DELETED
				return nil
			}

			// 500, In Creating , internal Error also happens
			if err == nil && resp.Error.Code == REQUEST_INVALID {
				return resource.RetryableError(fmt.Errorf("500 Retry"))
			}

			if err == nil && resp.Error.Code != REQUEST_COMPLETED {
				rds = resp.Result.DbInstanceAttributes
				rdsState = resp.Result.DbInstanceAttributes.InstanceStatus
				return resource.NonRetryableError(fmt.Errorf("rdsInstanceStatusRefreshFunc failed with %v", resp))
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

		return rds, rdsState, nil
	}
}

func rdsStatusWaiter(d *schema.ResourceData, meta interface{}, id string, pending, target []string) (err error) {

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    rdsInstanceStatusRefreshFunc(d, meta, id),
		Delay:      3 * time.Second,
		Timeout:    10 * time.Minute,
		MinTimeout: 1 * time.Second,
	}
	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in creatingDisk/Waiting disk,err message:%v", err)
	}
	return nil
}
