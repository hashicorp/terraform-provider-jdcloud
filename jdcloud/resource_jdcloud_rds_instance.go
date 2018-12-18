package jdcloud

import (
	"fmt"
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
			},
			"engine_version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
			},
			"charge_unit": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"charge_duration": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
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
	resp, err := rdsClient.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	if RDS_READY := waitForRDS(resp.Result.InstanceId, meta, RDS_READY); RDS_READY != nil {
		return RDS_READY
	}

	d.SetId(resp.Result.InstanceId)
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
	resp, err := rdsClient.DescribeInstanceAttributes(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSRead failed %s ", err.Error())
	}
	if resp.Error.Code == RESOURCE_NOT_FOUND || resp.Error.Code == REQUEST_INVALID {
		return nil
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

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
	d.Set("charge_mode", resp.Result.DbInstanceAttributes.Charge.ChargeMode)

	return nil
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
	return noOtherAttributesModified(d)
}

func resourceJDCloudRDSInstanceDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewDeleteInstanceRequest(config.Region, d.Id())
	resp, err := rdsClient.DeleteInstance(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDelete failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED && !invalidRequestDealing(resp.Error.Code, d, meta) {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	id := d.Id()
	d.SetId("")
	return waitForRDS(id, meta, RDS_DELETED)
}

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

func timeOut(currentTime int) bool {
	return int(time.Now().Unix())-currentTime > RDS_TIMEOUT
}
