package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	da "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	dc "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	dm "github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
	"time"
)

/*
  TODO Add these as reminder in webite
  1.Currently, the only way you can use cloud disk as [System Disk] is to
	select your region as "cn-east-2". Usually when you set your region as
	"cn-north-1", you can only use [local] disk instead,and the volume will
	be fixed to 40Gb

  2. disk type "premium-hdd" is currently out of stock, use [ssd] instead

  3. set no device as false to set up your data disk
*/

//----------------------------------------------------------------------------------- OTHERS

type vmLogger struct{ core.Logger }

func (l vmLogger) Log(level int, message ...interface{}) {
	log.Print(message...)
}

func stringAddr(v interface{}) *string {
	r := v.(string)
	return &r
}

func boolAddr(v interface{}) *bool {
	r := v.(bool)
	return &r
}

func GetStringAddr(d *schema.ResourceData, key string) *string {
	v := d.Get(key).(string)
	return &v
}

func GetIntAddr(d *schema.ResourceData, key string) *int {
	v := d.Get(key).(int)
	return &v
}

func QueryInstanceDetail(d *schema.ResourceData, m interface{}) (*apis.DescribeInstanceResponse, error) {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, d.Id())
	resp := new(apis.DescribeInstanceResponse)
	err := fmt.Errorf("")

	errRetry := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err = vmClient.DescribeInstance(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}

		if resp.Error.Code == 404 {
			resp.Result.Instance.Status = VM_DELETED
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}

	})

	if errRetry != nil {
		return nil, errRetry
	} else {
		return resp, err
	}
}

//----------------------------------------------------------------------------------- VM-RELATED

func waitForInstance(d *schema.ResourceData, m interface{}, expectedStatus string) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DescribeInstance(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED && resp.Result.Instance.Status == expectedStatus {
			return nil
		}

		if expectedStatus == "" && resp.Result.Instance.Status == expectedStatus {
			return nil
		}

		if connectionError(err) || resp.Result.Instance.Status != expectedStatus {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func StopVmInstance(d *schema.ResourceData, m interface{}) error {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewStopInstanceRequest(config.Region, d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.StopInstance(req)

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

func diskIdList(s *schema.Set) []string {

	i := []string{}

	for _, d := range s.List() {
		m := d.(map[string]interface{})
		i = append(i, m["disk_id"].(string))
	}
	return i
}

//----------------------------------------------------------------------------------- DISK-RELATED

func typeSetToDiskList(s *schema.Set) []vm.InstanceDiskAttachmentSpec {

	ds := []vm.InstanceDiskAttachmentSpec{}
	for _, d := range s.List() {

		c := dm.DiskSpec{}
		m := d.(map[string]interface{})
		i := vm.InstanceDiskAttachmentSpec{}

		if m["disk_category"] != "" {
			i.DiskCategory = stringAddr(m["disk_category"])
		}
		if m["auto_delete"] != "" {
			i.AutoDelete = boolAddr(m["auto_delete"])
		}
		if m["device_name"] != "" {
			i.DeviceName = stringAddr(m["device_name"])
		}
		if m["az"] != "" {
			c.Az = m["az"].(string)
		}
		if m["disk_name"] != "" {
			c.Name = m["disk_name"].(string)
		}
		if m["description"] != "" {
			c.Description = stringAddr(m["description"])
		}
		if m["disk_type"] != "" {
			c.DiskType = m["disk_type"].(string)
		}
		if m["disk_size_gb"] != "" {
			c.DiskSizeGB = m["disk_size_gb"].(int)
		}
		if m["snapshot_id"] != "" {
			c.SnapshotId = stringAddr(m["snapshot_id"])
		}

		i.CloudDiskSpec = &c
		ds = append(ds, i)
	}

	return ds
}

func diskListTypeCloud(d []vm.InstanceDiskAttachmentSpec) []vm.InstanceDiskAttachmentSpec {

	diskTypeCloud := DISKTYPE_CLOUD

	for _, item := range d {
		item.DiskCategory = &diskTypeCloud
	}

	return d
}

func cloudDiskStructIntoMap(ss []vm.InstanceDiskAttachment) []map[string]interface{} {

	ms := []map[string]interface{}{}

	for _, s := range ss {

		if s.Status != DISK_DETACHED {
			ms = append(ms, map[string]interface{}{
				"disk_category": s.DiskCategory,
				"auto_delete":   s.AutoDelete,
				"device_name":   s.DeviceName,
				"disk_id":       s.CloudDisk.DiskId,
				"az":            s.CloudDisk.Az,
				"disk_name":     s.CloudDisk.Name,
				"description":   s.CloudDisk.Description,
				"disk_type":     s.CloudDisk.DiskType,
				"disk_size_gb":  s.CloudDisk.DiskSizeGB,
			})
		}
	}

	return ms
}

func waitCloudDiskId(d *schema.ResourceData, m interface{}) error {

	resp, err := QueryInstanceDetail(d, m)

	if err != nil || resp.Error.Code != REQUEST_COMPLETED {
		return err
	}

	if errSet := d.Set("data_disk", cloudDiskStructIntoMap(resp.Result.Instance.DataDisks)); err != nil {
		return errSet
	}

	return nil
}

func performCloudDiskDetach(d *schema.ResourceData, m interface{}, set *schema.Set) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	detachList := diskIdList(set)

	// Keep sending all detach requests
	for _, id := range detachList {

		req := apis.NewDetachDiskRequest(config.Region, d.Id(), id)
		resp, err := vmClient.DetachDisk(req)

		if err != nil || resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] performCloudDiskDetach Failed, reasons: %s,%s", err.Error(), resp.Error.Message)
		}
	}

	// Wait until all requests completed
	for _, id := range detachList {
		err := waitForDiskAttaching(d, m, d.Id(), id, DISK_DETACHED)
		if err != nil {
			return fmt.Errorf("[ERROR] performCloudDiskDetach Failed, reasons: %s", err.Error())
		}
	}

	return nil
}

func performCloudDiskAttach(d *schema.ResourceData, m interface{}, set *schema.Set) error {

	ids, err := performNewDiskCreate(d, m, diskListTypeCloud(typeSetToDiskList(set)))

	if err != nil {
		return err
	}

	if err := performNewDiskAttach(d, m, ids); err != nil {
		return err
	}

	return nil
}

func performNewDiskCreate(d *schema.ResourceData, m interface{}, diskSpecsCloud []vm.InstanceDiskAttachmentSpec) ([]string, error) {

	ids := []string{}
	config := m.(*JDCloudConfig)
	diskClient := dc.NewDiskClient(config.Credential)

	for _, item := range diskSpecsCloud {

		req := da.NewCreateDisksRequest(
			config.Region,
			item.CloudDiskSpec,
			MAX_DISK_COUNT,
			diskClientTokenDefault())
		resp, err := diskClient.CreateDisks(req)

		if err != nil {
			return nil, fmt.Errorf("[ERROR] performCloudDiskAttach Failed, reasons: %s", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return nil, fmt.Errorf("[ERROR] performCloudDiskAttach failed  code:%d staus:%s message:%s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		ids = append(ids, resp.Result.DiskIds[0])
	}

	for _, diskId := range ids {

		err := waitForDisk(d, m, diskId, DISK_AVAILABLE)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] performCloudDiskAttach Failed, reasons: %s", err.Error())
		}
	}

	return ids, nil
}

func performNewDiskAttach(d *schema.ResourceData, m interface{}, ids []string) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)

	for _, i := range ids {

		req := apis.NewAttachDiskRequest(config.Region, d.Id(), i)
		resp, err := vmClient.AttachDisk(req)

		if err != nil {
			return fmt.Errorf("[ERROR] performNewDiskAttach failed %s ", err.Error())
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] performNewDiskAttach  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}

	for _, i := range ids {

		if errAttaching := waitForDiskAttaching(d, m, d.Id(), i, DISK_ATTACHED); errAttaching != nil {
			return fmt.Errorf("[ERROR] failed in attaching disk,reasons: %s", errAttaching.Error())
		}
	}

	return nil
}

//----------------------------------------------------------------------------------- RESOURCE

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
			"disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

	return &schema.Resource{
		Create: resourceJDCloudInstanceCreate,
		Read:   resourceJDCloudInstanceRead,
		Update: resourceJDCloudInstanceUpdate,
		Delete: resourceJDCloudInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"az": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"key_names": { //Only one key pair name is supported
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"primary_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
				MaxItems: MAX_SECURITY_GROUP_COUNT,
			},

			"network_interface_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"secondary_ips": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},
			"secondary_ip_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"sanity_check": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"elastic_ip_bandwidth_mbps": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"elastic_ip_provider": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"system_disk": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem:     diskSchema,
				MaxItems: MAX_SYSDISK_COUNT,
				ForceNew: true,
			},
			"data_disk": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Elem:     diskSchema,
			},
		},
	}
}

func resourceJDCloudInstanceCreate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
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

	if _, ok := d.GetOk("system_disk"); ok {
		spec.SystemDisk = &(typeSetToDiskList(d.Get("system_disk").(*schema.Set))[0])
	}

	if _, ok := d.GetOk("data_disk"); ok {
		d := typeSetToDiskList(d.Get("data_disk").(*schema.Set))
		// This step is introduced since disk type uses cloud disk only
		spec.DataDisks = diskListTypeCloud(d)
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

	if _, ok := d.GetOk("secondary_ips"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SecondaryIpAddresses = typeSetToStringArray(d.Get("secondary_ips").(*schema.Set))
	}

	if _, ok := d.GetOk("secondary_ip_count"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SecondaryIpCount = GetIntAddr(d, "secondary_ip_count")
	}

	if _, ok := d.GetOk("sanity_check"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SanityCheck = GetIntAddr(d, "sanity_check")
	}

	if _, ok := d.GetOk("security_group_ids"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SecurityGroups = typeSetToStringArray(d.Get("security_group_ids").(*schema.Set))
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
	req.SetMaxCount(MAX_VM_COUNT)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.CreateInstances(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.InstanceIds[0])
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

	errCreating := waitForInstance(d, m, VM_RUNNING)
	if errCreating != nil {
		d.SetId("")
		return errCreating
	}

	d.SetPartial("az")
	d.SetPartial("image_id")
	d.SetPartial("password")
	d.SetPartial("subnet_id")
	d.SetPartial("key_names")
	d.SetPartial("primary_ip")
	d.SetPartial("system_disk")
	d.SetPartial("description")
	d.SetPartial("instance_type")
	d.SetPartial("instance_name")
	d.SetPartial("secondary_ips")
	d.SetPartial("secondary_ip_count")
	d.SetPartial("security_group_ids")
	d.SetPartial("elastic_ip_provider")
	d.SetPartial("network_interface_name")
	d.SetPartial("elastic_ip_bandwidth_mbps")

	if errDataDisk := waitCloudDiskId(d, m); errDataDisk != nil {
		d.SetId("")
		return errDataDisk
	}

	d.SetPartial("data_disk")
	d.Partial(false)
	return nil
}

func resourceJDCloudInstanceRead(d *schema.ResourceData, m interface{}) error {

	vmInstanceDetail, err := QueryInstanceDetail(d, m)

	if err != nil {
		return fmt.Errorf("query vm instance fail: %s", err)
	}
	if vmInstanceDetail.Result.Instance.Status == VM_DELETED || vmInstanceDetail.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
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

	if errSet := d.Set("security_group_ids", vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.SecurityGroups); err != nil {
		return fmt.Errorf("[ERROR] Failed in setting Sg Id LIST, reasons:%s", errSet.Error())
	}

	if errSet := d.Set("secondary_ips", vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.SecondaryIps); err != nil {
		return fmt.Errorf("[ERROR] Failed in setting secondary ip LIST, reasons:%s", errSet.Error())
	}

	if errSet := d.Set("data_disk", cloudDiskStructIntoMap(vmInstanceDetail.Result.Instance.DataDisks)); err != nil {
		return fmt.Errorf("[ERROR] Failed in setting data_disk, reasons:%s", errSet.Error())
	}

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

		if err := StopVmInstance(d, m); err != nil {
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
			if err := StopVmInstance(d, m); err != nil {
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

	if d.HasChange("data_disk") {

		log.Printf("[WARN] Trying to modify the property of data disk, leads to NEW DISK REBUILDING")

		pInterface, cInterface := d.GetChange("data_disk")
		p := pInterface.(*schema.Set)
		c := cInterface.(*schema.Set)
		i := p.Intersection(c)

		if err := performCloudDiskDetach(d, m, p.Difference(i)); len(typeSetToDiskList(p.Difference(i))) != 0 && err != nil {
			return err
		}
		if err := performCloudDiskAttach(d, m, c.Difference(i)); len(typeSetToDiskList(c.Difference(i))) != 0 && err != nil {
			return err
		}

		d.SetPartial("data_disk")
	}

	d.Partial(false)
	return nil
}

func resourceJDCloudInstanceDelete(d *schema.ResourceData, m interface{}) error {
	vmInstanceDetail, err := QueryInstanceDetail(d, m)
	if err != nil {
		return err
	}

	if vmInstanceDetail.Result.Instance.Status == VM_RUNNING {
		if err := StopVmInstance(d, m); err != nil {
			return err
		}
		if err := waitForInstance(d, m, VM_STOPPED); err != nil {
			return err
		}
	}

	errDel := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := DeleteVmInstance(d, m)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			return nil
		}
		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
	if errDel != nil {
		return errDel
	}

	if err := waitForInstance(d, m, VM_DELETED); err != nil {
		return err
	}
	d.SetId("")
	return nil
}
