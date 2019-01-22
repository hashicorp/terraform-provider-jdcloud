package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	rds "github.com/jdcloud-api/jdcloud-sdk-go/services/rds/models"
	"log"
	"regexp"
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
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"instance_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
			"rds_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
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
	d.Partial(true)

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

	rdsClient := client.NewRdsClient(config.Credential)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.CreateInstance(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.InstanceId)
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
	reqStatus := apis.NewDescribeInstanceAttributesRequest(config.Region, d.Id())
	timeOut := resource.Retry(10*time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.DescribeInstanceAttributes(reqStatus)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED && resp.Result.DbInstanceAttributes.InstanceStatus == RDS_READY {
			return nil
		}

		// Process way too faster than network, sometimes RDS instance has been created and
		// shown to be "CREATING", and appears to be your Instance not exsits with err code 400
		if connectionError(err) || resp.Error.Code == REQUEST_INVALID ||
			resp.Result.DbInstanceAttributes.InstanceStatus != RDS_READY {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}

	})
	if timeOut != nil {
		return timeOut
	}

	d.Set("rds_id", d.Id())
	d.SetPartial("instance_name")
	d.SetPartial("engine")
	d.SetPartial("engine_version")
	d.SetPartial("instance_class")
	d.SetPartial("instance_storage_gb")
	d.SetPartial("az")
	d.SetPartial("vpc_id")
	d.SetPartial("vpc_id")
	d.SetPartial("rds_id")
	d.SetPartial("instance_port")
	d.SetPartial("connection_mode")
	d.SetPartial("charge_mode")
	d.SetPartial("charge_unit")
	d.SetPartial("charge_duration")

	// This step is added since "domain_name" is needed but only available through reading
	if err := resourceJDCloudRDSInstanceRead(d, meta); err == nil {
		d.SetPartial("internal_domain_name")
		d.SetPartial("public_domain_name")
	} else {
		log.Printf("Resource Created but failed to load its name,reasons: %s", err.Error())
	}

	d.Partial(false)
	return nil
}

func resourceJDCloudRDSInstanceRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	req := apis.NewDescribeInstanceAttributesRequest(config.Region, d.Id())
	rdsClient := client.NewRdsClient(config.Credential)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {

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

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	if d.HasChange("instance_name") {
		req := apis.NewSetInstanceNameRequest(config.Region, d.Id(), d.Get("instance_name").(string))
		resp, err := rdsClient.SetInstanceName(req)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed %s ", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
		d.SetPartial("instance_name")
	}

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
		d.SetPartial("instance_class")
		d.SetPartial("instance_storage_gb")
	}

	d.Partial(false)
	return nil
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
	reqStatus := apis.NewDescribeInstanceAttributesRequest(config.Region, d.Id())
	return resource.Retry(10*time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.DescribeInstanceAttributes(reqStatus)
		if resp.Result.DbInstanceAttributes.InstanceStatus == RDS_DELETED {
			d.SetId("")
			return nil
		}

		// Not suppose to happen, Just in case
		if resp.Error.Code == RESOURCE_NOT_FOUND {
			d.SetId("")
			return nil
		}

		if connectionError(err) || resp.Error.Code == REQUEST_INVALID || resp.Result.DbInstanceAttributes.InstanceStatus != RDS_DELETED {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}

	})
}

// DISCARDED FUNCTIONS - expected to be deleted in the near future :-)

func waitForRDS(id string, meta interface{}, expectedStatus string) error {

	t := int(time.Now().Unix())
	failCount := 0
	config := meta.(*JDCloudConfig)
	c := client.NewRdsClient(config.Credential)
	req := apis.NewDescribeInstanceAttributesRequest(config.Region, id)

	for {

		resp, err := c.DescribeInstanceAttributes(req)
		if resp.Result.DbInstanceAttributes.InstanceStatus == expectedStatus {
			return nil
		}

		if timeOut(t) {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed, timeout")
		}

		if !errDealing(err, &failCount) {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSWait, Tolerance Exceeded failed %s ", err.Error())
		}
	}

}

func errDealing(err error, failCount *int) bool {

	if err != nil {

		if s, _ := regexp.MatchString(CONNECT_FAILED, err.Error()); !s || *failCount > MAX_RECONNECT_COUNT {
			return false
		}

		*failCount++
		time.Sleep(10 * time.Second)
		return true
	}

	*failCount = 0
	return true
}

func timeOut(currentTime int) bool {
	return int(time.Now().Unix())-currentTime > RDS_TIMEOUT
}

func noOtherAttributesModified(d *schema.ResourceData) error {
	remainingAttr := []string{"charge_duration", "charge_unit", "charge_mode", "connection_mode",
		"internal_domain_name", "rds_id", "subnet_id", "vpc_id", "az", "instance_name",
		"engine", "engine_version"}
	attrModified := false
	for _, attr := range remainingAttr {
		if d.HasChange(attr) {
			origin, _ := d.GetChange(attr)
			d.Set(attr, origin)
			attrModified = true
		}
	}
	if attrModified {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed , Other attributes cannot be modified")
	} else {
		return nil
	}
}

func invalidRequestDealing(c int, d *schema.ResourceData, m interface{}) bool {

	if c != REQUEST_INVALID {
		return false
	}

	time.Sleep(10 * time.Second)
	config := m.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewDeleteInstanceRequest(config.Region, d.Id())
	resp, err := rdsClient.DeleteInstance(req)

	if err != nil || resp.Error.Code != REQUEST_COMPLETED {
		return false
	}

	return true
}
