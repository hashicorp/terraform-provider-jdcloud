package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	disk "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
	"time"
)

/*
  Reminder:
  1.Currently, the only way you can use cloud disk as [System Disk] is to
	select your region as "cn-east-2". Usually when you set your region as
	"cn-north-1", you can only use [local] disk instead,and the volume will
	be fixed to 40Gb

  2. disk type "premium-hdd" is currently out of stock, use [ssd] instead

  3. set no device as false to set up your data disk
*/

func resourceJDCloudInstance() *schema.Resource {
	diskSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"disk_category": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_delete": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"device_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"no_device": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"az": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disk_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disk_size_gb": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}

	return &schema.Resource{
		Create: resourceJDCloudInstanceCreate,
		Read:   resourceJDCloudInstanceRead,
		Update: resourceJDCloudInstanceUpdate,
		Delete: resourceJDCloudInstanceDelete,

		Schema: map[string]*schema.Schema{
			"az": {
				Type:     schema.TypeString,
				Required: true,
			},
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"key_names": { //Only one key pair name is supported
				Type:     schema.TypeString,
				Optional: true,
			},

			"primary_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"network_interface_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secondary_ips": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"secondary_ip_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"sanity_check": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"elastic_ip_bandwidth_mbps": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"elastic_ip_provider": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"system_disk": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     diskSchema,
			},
			"data_disk": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     diskSchema,
			},
		},
	}
}

func GetStringAddr(d *schema.ResourceData, key string) *string {
	v := d.Get(key).(string)
	return &v
}

func GetIntAddr(d *schema.ResourceData, key string) *int {
	v := d.Get(key).(int)
	return &v
}

func InterfaceToStringArray(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, v.(string))
	}
	return vs
}

func waitForInstance(d *schema.ResourceData, m interface{}, vmStatus string) error {
	currentTime := int(time.Now().Unix())

	for {
		if int(time.Now().Unix())-currentTime >= DefaultTimeout {
			return errors.New("create vm instance timeout")
		}
		vmInstanceDetail, err := QueryInstanceDetail(d, m)
		if err != nil {
			return errors.New("query vm instance detail fail")
		}
		if vmInstanceDetail.Result.Instance.Status != vmStatus {
			continue
		}

		return nil
	}
}

func QueryInstanceDetail(d *schema.ResourceData, m interface{}) (*apis.DescribeInstanceResponse, error) {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, d.Id())
	resp, err := vmClient.DescribeInstance(req)
	if resp.Error.Code == 404 {
		resp.Result.Instance.Status = VM_DELETED
	}
	return resp, err
}

func StopVmInstance(d *schema.ResourceData, m interface{}) (*apis.StopInstanceResponse, error) {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewStopInstanceRequest(config.Region, d.Id())
	resp, err := vmClient.StopInstance(req)
	return resp, err
}

func StartVmInstance(d *schema.ResourceData, m interface{}) (*apis.StartInstanceResponse, error) {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewStartInstanceRequest(config.Region, d.Id())
	resp, err := vmClient.StartInstance(req)
	return resp, err
}

func DeleteVmInstance(d *schema.ResourceData, m interface{}) (*apis.DeleteInstanceResponse, error) {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDeleteInstanceRequest(config.Region, d.Id())
	resp, err := vmClient.DeleteInstance(req)
	return resp, err
}

type vmLogger struct{ core.Logger }

func (l vmLogger) Log(level int, message ...interface{}) {
	log.Printf("[VM]", message...)
}

func resourceJDCloudInstanceCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	logger := vmLogger{}
	vmClient.SetLogger(logger)

	spec := vm.InstanceSpec{
		Az:           GetStringAddr(d, "az"),
		InstanceType: GetStringAddr(d, "instance_type"),
		ImageId:      GetStringAddr(d, "image_id"),
		Name:         d.Get("instance_name").(string),
		PrimaryNetworkInterface: &vm.InstanceNetworkInterfaceAttachmentSpec{
			NetworkInterface: &vpc.NetworkInterfaceSpec{SubnetId: d.Get("subnet_id").(string), Az: GetStringAddr(d, "az")},
		},
	}

	if v, ok := d.GetOk("system_disk"); ok {
		systemDisk := v.([]interface{})
		if len(systemDisk) > 0 {
			spec.SystemDisk = newDiskSpecFromSchema(systemDisk[0].(map[string]interface{}))
		}
	}

	if v, ok := d.GetOk("data_disk"); ok {
		dataDisk := []vm.InstanceDiskAttachmentSpec{}
		for _, v := range v.([]interface{}) {
			spec := newDiskSpecFromSchema(v.(map[string]interface{}))
			dataDisk = append(dataDisk, *spec)
		}
		spec.DataDisks = dataDisk
	}

	if _, ok := d.GetOk("description"); ok {
		spec.Description = GetStringAddr(d, "description")
	}

	if _, ok := d.GetOk("password"); ok {
		spec.Password = GetStringAddr(d, "password")
	}

	if _, ok := d.GetOk("key_names"); ok {
		spec.KeyNames = []string{d.Get("key_names").(string)}
	}

	if _, ok := d.GetOk("primary_ip"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.PrimaryIpAddress = GetStringAddr(d, "primary_ip")
	}
	if _, ok := d.GetOk("network_interface_name"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.NetworkInterfaceName = GetStringAddr(d, "network_interface_name")
	}
	if v, ok := d.GetOk("secondary_ips"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SecondaryIpAddresses = InterfaceToStringArray(v.([]interface{}))
	}
	if _, ok := d.GetOk("secondary_ip_count"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SecondaryIpCount = GetIntAddr(d, "secondary_ip_count")
	}
	if _, ok := d.GetOk("sanity_check"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SanityCheck = GetIntAddr(d, "sanity_check")
	}
	if v, ok := d.GetOk("security_group_ids"); ok {
		sgList := InterfaceToStringArray(v.([]interface{}))
		if len(sgList) > DefaultSecurityGroupsMax {
			return fmt.Errorf("the maximum allowed number of security_group_ids is %d", DefaultSecurityGroupsMax)
		}
		spec.PrimaryNetworkInterface.NetworkInterface.SecurityGroups = sgList
	}

	if v, ok := d.GetOk("elastic_ip_bandwidth_mbps"); ok {
		if spec.ElasticIp == nil {
			spec.ElasticIp = &vpc.ElasticIpSpec{}
		}
		spec.ElasticIp.BandwidthMbps = v.(int)
	}
	if v, ok := d.GetOk("elastic_ip_provider"); ok {
		if spec.ElasticIp == nil {
			spec.ElasticIp = &vpc.ElasticIpSpec{}
		}
		spec.ElasticIp.Provider = v.(string)
	}

	req := apis.NewCreateInstancesRequest(config.Region, &spec)
	req.SetMaxCount(1)

	resp, err := vmClient.CreateInstances(req)
	if err != nil {
		return err
	} else if resp.Error.Code != 0 {
		return fmt.Errorf("Create vm instance failed: %s", resp.Error)
	}

	d.SetId(resp.Result.InstanceIds[0])
	return waitForInstance(d, m, VM_RUNNING)
}

func resourceJDCloudInstanceRead(d *schema.ResourceData, m interface{}) error {
	vmInstanceDetail, err := QueryInstanceDetail(d, m)
	if err != nil {
		if vmInstanceDetail.Result.Instance.Status == VM_DELETED {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("query vm instance fail: %s", err)
	}

	d.Set("instance_name", vmInstanceDetail.Result.Instance.InstanceName)
	d.Set("image_id", vmInstanceDetail.Result.Instance.ImageId)
	d.Set("instance_type", vmInstanceDetail.Result.Instance.InstanceType)
	d.Set("password", d.Get("password"))
	d.Set("description", vmInstanceDetail.Result.Instance.Description)
	d.Set("subnet_id", vmInstanceDetail.Result.Instance.SubnetId)
	d.Set("primary_ip", vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.PrimaryIp)
	d.Set("elastic_ip", vmInstanceDetail.Result.Instance.ElasticIpAddress)
	d.Set("key_names", vmInstanceDetail.Result.Instance.KeyNames)
	d.Set("security_group_ids", vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.SecurityGroups)
	return nil
}

func resourceJDCloudInstanceUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)

	if d.HasChange("instance_name") || d.HasChange("description") {
		req := apis.ModifyInstanceAttributeRequest{
			RegionId:    config.Region,
			InstanceId:  d.Id(),
			Name:        GetStringAddr(d, "instance_name"),
			Description: GetStringAddr(d, "description"),
		}
		if _, err := vmClient.ModifyInstanceAttribute(&req); err != nil {
			return err
		}
		d.SetPartial("instance_name")
		d.SetPartial("description")
	}

	if d.HasChange("password") {

		if _, err := StopVmInstance(d, m); err != nil {
			return fmt.Errorf("stop instance got error:%s", err)
		}
		req := apis.ModifyInstancePasswordRequest{
			RegionId:   config.Region,
			InstanceId: d.Id(),
			Password:   d.Get("password").(string),
		}

		if _, err := vmClient.ModifyInstancePassword(&req); err != nil {
			return err
		}

		vmInstanceDetail, err := QueryInstanceDetail(d, m)
		if err != nil {
			return err
		}

		if vmInstanceDetail.Result.Instance.Status == VM_RUNNING {
			if _, err := StopVmInstance(d, m); err != nil {
				return fmt.Errorf("stop instance fail: %s", err)
			}
		}

		if err := waitForInstance(d, m, VM_STOPPED); err != nil {
			return fmt.Errorf("query stopped instance fail:%s", err)
		}

		if _, err := StartVmInstance(d, m); err != nil {
			return fmt.Errorf("start instance fail: %s", err)
		}

		if err := waitForInstance(d, m, VM_RUNNING); err != nil {
			return fmt.Errorf("query running instance fail:%s", err)
		}

		d.SetPartial("password")
	}

	d.Partial(false)
	return nil
}

func resourceJDCloudInstanceDelete(d *schema.ResourceData, m interface{}) error {

	vmInstanceDetail, err := QueryInstanceDetail(d, m)
	if err != nil {
		return fmt.Errorf("query instance fail: %s", err)
	}

	if vmInstanceDetail.Result.Instance.Status == VM_RUNNING {
		if _, err := StopVmInstance(d, m); err != nil {
			return fmt.Errorf("stop instance fail: %s", err)
		}
		if err := waitForInstance(d, m, VM_STOPPED); err != nil {
			return fmt.Errorf("query stopped instance fail: %s", err)
		}
	}

	if _, err := DeleteVmInstance(d, m); err != nil {
		return fmt.Errorf("delete instance fail: %s", err)
	}

	if err := waitForInstance(d, m, VM_DELETED); err != nil {
		return fmt.Errorf("query deleted instance fail: %s", err)
	}

	d.SetId("")

	return nil
}

func stringAddr(v interface{}) *string {
	r := v.(string)
	return &r
}

func boolAddr(v interface{}) *bool {
	r := v.(bool)
	return &r
}

func newDiskSpecFromSchema(m map[string]interface{}) *vm.InstanceDiskAttachmentSpec {
	spec := &vm.InstanceDiskAttachmentSpec{}
	if v, ok := m["disk_category"]; ok {
		spec.DiskCategory = stringAddr(v)
	}
	if v, ok := m["auto_delete"]; ok {
		spec.AutoDelete = boolAddr(v)
	}
	if v, ok := m["device_name"]; ok {
		spec.DeviceName = stringAddr(v)
	}
	if v, ok := m["no_device"]; ok {
		spec.NoDevice = boolAddr(v)
	}

	cloudDiskScheme := []string{"az", "disk_name", "description", "disk_type", "disk_size_gb", "snapshot_id"}
	for _, v := range cloudDiskScheme {
		_, ok := m[v]
		if ok && spec.CloudDiskSpec == nil {
			spec.CloudDiskSpec = &disk.DiskSpec{}
		}
	}

	if v, ok := m["az"]; ok {
		spec.CloudDiskSpec.Az = v.(string)
	}
	if v, ok := m["disk_name"]; ok {
		spec.CloudDiskSpec.Name = v.(string)
	}
	if v, ok := m["description"]; ok {
		spec.CloudDiskSpec.Description = stringAddr(v)
	}
	if v, ok := m["disk_type"]; ok {
		spec.CloudDiskSpec.DiskType = v.(string)
	}
	if v, ok := m["disk_size_gb"]; ok {
		spec.CloudDiskSpec.DiskSizeGB = v.(int)
	}
	if v, ok := m["snapshot_id"]; ok {
		spec.CloudDiskSpec.SnapshotId = stringAddr(v)
	}
	return spec
}
