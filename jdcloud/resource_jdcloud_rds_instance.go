package jdcloud

import (
	"fmt"
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
const (
	RDSTimeout = 300
	RDSReady   = "RUNNING"
	RDSDeleted = ""
	Tolerance  = 3
)

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
		chargeSpec = &models.ChargeSpec{&chargeMode, GetStringAddr(d, "charge_unit"), GetIntAddr(d, "charge_duration")}
	} else {
		chargeSpec = &models.ChargeSpec{&chargeMode, nil, nil}
	}
	req := apis.NewCreateInstanceRequest(
		config.Region,
		&rds.DBInstanceSpec{
			GetStringAddr(d, "instance_name"),
			d.Get("engine").(string),
			d.Get("engine_version").(string),
			d.Get("instance_class").(string),
			d.Get("instance_storage_gb").(int),
			[]string{d.Get("az").(string)},
			d.Get("vpc_id").(string),
			d.Get("subnet_id").(string),
			chargeSpec,
		},
	)
	rdsClient := client.NewRdsClient(config.Credential)
	resp, err := rdsClient.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed %s ", err.Error())
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.SetId(resp.Result.InstanceId)
	d.Set("rds_id", d.Id())
	if rdsReady := waitForRDS(d.Id(), meta, RDSReady); rdsReady != nil {
		return rdsReady
	}
	// This step is added since "domain_name" is needed but only available through reading
	return resourceJDCloudRDSInstanceRead(d, meta)
}
func resourceJDCloudRDSInstanceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*JDCloudConfig)
	req := apis.NewDescribeInstanceAttributesRequest(config.Region, d.Id())
	rdsClient := client.NewRdsClient(config.Credential)
	resp, err := rdsClient.DescribeInstanceAttributes(req)
	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSRead failed %s ", err.Error())
	}
	if resp.Error.Code == 404 || resp.Error.Code == 400 {
		return nil
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.Set("instance_name", resp.Result.DbInstanceAttributes.InstanceName)
	d.Set("instance_class", resp.Result.DbInstanceAttributes.InstanceClass)
	d.Set("instance_storage_gb", resp.Result.DbInstanceAttributes.InstanceStorageGB)
	d.Set("internal_domain_name", resp.Result.DbInstanceAttributes.InternalDomainName)
	d.Set("public_domain_name", resp.Result.DbInstanceAttributes.PublicDomainName)
	d.Set("instance_port", resp.Result.DbInstanceAttributes.InstancePort)
	d.Set("connection_mode", resp.Result.DbInstanceAttributes.ConnectionMode)
	return nil
}
func resourceJDCloudRDSInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	if d.HasChange("instance_name") {
		req := apis.NewSetInstanceNameRequest(config.Region, d.Id(), d.Get("instance_name").(string))
		resp, err := rdsClient.SetInstanceName(req)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed %s ", err.Error())
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}
	// Currently you can not degrade your configuration, only upgrade them is allowed
	if d.HasChange("instance_class") || d.HasChange("instance_storage_gb") {
		req := apis.NewModifyInstanceSpecRequest(config.Region, d.Id(), d.Get("instance_class").(string), d.Get("instance_storage_gb").(int))
		resp, err := rdsClient.ModifyInstanceSpec(req)
		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed %s ", err.Error())
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}
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
	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	id := d.Id()
	d.SetId("")
	return waitForRDS(id, meta, RDSDeleted)
}
func waitForRDS(id string, meta interface{}, expectedStatus string) error {
	currentTime := int(time.Now().Unix())
	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewDescribeInstanceAttributesRequest(config.Region, id)
	connectFailedCount := 0
	for {
		time.Sleep(time.Second * 10)
		resp, err := rdsClient.DescribeInstanceAttributes(req)
		if resp.Result.DbInstanceAttributes.InstanceStatus == expectedStatus {
			return nil
		}
		if int(time.Now().Unix())-currentTime > RDSTimeout {
			return fmt.Errorf("[ERROR] resourceJDCloudRDSCreate failed, timeout")
		}
		if err != nil {
			if connectFailedCount > Tolerance {
				return fmt.Errorf("[ERROR] resourceJDCloudRDSWait, Tolerrance Exceeded failed %s ", err.Error())
			}
			connectFailedCount++
			continue
		} else {
			connectFailedCount = 0
		}
	}
	return nil
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
