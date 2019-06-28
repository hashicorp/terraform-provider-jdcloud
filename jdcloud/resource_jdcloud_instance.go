package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
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

func intAddr(v interface{}) *int {
	r := v.(int)
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

func QueryInstanceDetail(d *schema.ResourceData, m interface{}, instanceId string) (r *apis.DescribeInstanceResponse, e error) {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, instanceId)
	e = resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DescribeInstance(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			r = resp
			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND {
			resp.Result.Instance.Status = VM_DELETED
			r = resp
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}

	})
	return r, e
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

// Level 2 -> Based on instanceStatusWaiter

func StopVmInstance(d *schema.ResourceData, m interface{}, instanceId string) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewStopInstanceRequest(config.Region, instanceId)

	return resource.Retry(time.Minute, func() *resource.RetryError {

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

	//if e != nil {
	//	return e
	//}
	//return instanceStatusWaiter(d, m, d.Id(), []string{VM_RUNNING, VM_STOPPING}, []string{VM_STOPPED,VM_STOPPED_2})
}

func StartVmInstance(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewStartInstanceRequest(config.Region, d.Id())

	e := resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := vmClient.StartInstance(req)

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
	return instanceStatusWaiter(d, m, d.Id(), []string{VM_STOPPED, VM_STARTING}, []string{VM_RUNNING})
}

func DeleteVmInstance(d *schema.ResourceData, m interface{}, id string) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)

	err := resource.Retry(time.Minute, func() *resource.RetryError {

		req := apis.NewDeleteInstanceRequest(config.Region, id)
		resp, err := vmClient.DeleteInstance(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
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

func diskIdList(s *schema.Set) []string {

	i := []string{}

	for _, d := range s.List() {
		m := d.(map[string]interface{})
		i = append(i, m["disk_id"].(string))
	}
	return i
}

// Used to refresh vm status level 0
func instanceStatusRefreshFunc(d *schema.ResourceData, meta interface{}, vmId string) resource.StateRefreshFunc {

	return func() (vmItem interface{}, vmStatus string, e error) {

		err := resource.Retry(time.Minute, func() *resource.RetryError {
			config := meta.(*JDCloudConfig)
			c := client.NewVmClient(config.Credential)
			req := apis.NewDescribeInstanceRequest(config.Region, vmId)
			resp, err := c.DescribeInstance(req)
			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				vmItem = resp.Result.Instance
				vmStatus = resp.Result.Instance.Status
				return nil
			}

			if connectionError(err) {
				return resource.RetryableError(err)
			} else {
				vmItem = resp.Result.Instance
				vmStatus = resp.Result.Instance.Status
				return resource.NonRetryableError(err)
			}

		})

		if err != nil {
			return nil, "", err
		}
		return vmItem, vmStatus, nil
	}
}

// Used to refresh until instance reached expected status level 1 -> Based on instanceStatusRefreshFunc
func instanceStatusWaiter(d *schema.ResourceData, meta interface{}, id string, pending, target []string) (err error) {

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    instanceStatusRefreshFunc(d, meta, id),
		Delay:      3 * time.Second,
		Timeout:    2 * time.Minute,
		MinTimeout: 1 * time.Second,
	}
	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("[E] Failed in instanceStatusWaiter/Waiting to reach expected status ,err message:%v", err)
	}
	return nil
}

// Level 2~3  delete a specified instance
func deleteInstance(d *schema.ResourceData, m interface{}, instanceId string) error {

	// Stop VM
	err := StopVmInstance(d, m, instanceId)
	if err != nil {
		return fmt.Errorf("[E] deleteInstance - InstanceId=%s - Send request to stop fail:%v", instanceId, err)
	}

	// Wait until stopped
	err = instanceStatusWaiter(d, m, instanceId, []string{VM_RUNNING, VM_STOPPING}, []string{VM_STOPPED, VM_STOPPED_2})
	if err != nil {
		return fmt.Errorf("[E] deleteInstance - InstanceId=%s - Can not make it stop :%v", instanceId, err)
	}

	// Delete VM
	err = DeleteVmInstance(d, m, instanceId)
	if err != nil {
		return fmt.Errorf("[E] deleteInstance - InstanceId=%s - Can not send requests to delete :%v", instanceId, err)
	}

	// Wait until deleted
	if err = instanceStatusWaiter(d, m, instanceId, []string{VM_RUNNING, VM_STOPPING, VM_DELETING}, []string{VM_DELETED}); err != nil {
		return fmt.Errorf("[E] deleteInstance - InstanceId=%s - Can not wait it delete :%v", instanceId, err)
	}

	return nil
}

// Level 2~3 delete some instances
func deleteInstances(d *schema.ResourceData, m interface{}, instanceIds []string) error {

	// Stop
	for _, i := range instanceIds {
		if e := StopVmInstance(d, m, i); e != nil {
			return fmt.Errorf("[E] deleteInstances - Send request to stop fail:%s", e)
		}
	}

	//Wait until stopped
	for _, i := range instanceIds {
		if err := instanceStatusWaiter(d, m, i, []string{VM_RUNNING, VM_STOPPING}, []string{VM_STOPPED, VM_STOPPED_2}); err != nil {
			return fmt.Errorf("[E] deleteInstances - Can not make it stop :%s", err)
		}
	}

	// Delete
	for _, i := range instanceIds {
		if e := DeleteVmInstance(d, m, i); e != nil {
			return fmt.Errorf("[E] deleteInstances - Can not send requests to delete:%v", e)
		}
	}

	// Wait until deleted
	for _, i := range instanceIds {
		if e := instanceStatusWaiter(d, m, i, []string{VM_RUNNING, VM_STOPPING, VM_DELETING}, []string{VM_DELETED}); e != nil {
			return fmt.Errorf("[E] deleteInstances - Can not wait it delete %v", e)
		}
	}
	return nil
}

//----------------------------------------------------------------------------------- DISK-RELATED

func typeListToDiskList(s []interface{}) []vm.InstanceDiskAttachmentSpec {

	ds := []vm.InstanceDiskAttachmentSpec{}
	for _, d := range s {

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

			if s.CloudDisk.DiskSizeGB != 0 {

				// Cloud-Disk
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
			} else {

				// Local-Disk
				ms = append(ms, map[string]interface{}{
					"disk_category": s.DiskCategory,
					"auto_delete":   s.AutoDelete,
					"device_name":   s.DeviceName,
					"disk_id":       "",
					"az":            "",
					"disk_name":     "",
					"description":   "",
					"disk_type":     s.LocalDisk.DiskType,
					"disk_size_gb":  s.LocalDisk.DiskSizeGB,
				})
			}

		}
	}

	return ms
}

func sysDiskStructIntoMap(s vm.InstanceDiskAttachment) map[string]interface{} {

	return map[string]interface{}{
		"disk_category": s.DiskCategory,
		"auto_delete":   s.AutoDelete,
		"device_name":   s.DeviceName,
		"disk_size_gb":  s.LocalDisk.DiskSizeGB,
	}
}

func waitCloudDiskId(d *schema.ResourceData, m interface{}) error {

	resp, err := QueryInstanceDetail(d, m, d.Id())

	if err != nil || resp.Error.Code != REQUEST_COMPLETED {
		return err
	}

	if errSet := d.Set("data_disk", cloudDiskStructIntoMap(resp.Result.Instance.DataDisks)); err != nil {
		return errSet
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
				ForceNew: true,
			},
			"auto_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"device_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"az": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"disk_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"disk_size_gb": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  40,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"disk_id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
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

			// You set : secondary_ip_count (Optional)
			// You got : ip_addresses (Computed)
			"secondary_ip_count": {
				Type:      schema.TypeInt,
				Optional:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"ip_addresses": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem:     diskSchema,
				MaxItems: MAX_SYSDISK_COUNT,
				ForceNew: true,
			},
			"data_disk": {
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
				Elem:     diskSchema,
			},
		},
	}
}

func resourceJDCloudInstanceCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	logger := vmLogger{}
	vmClient.SetLogger(logger)

	// Preparing necessary parameters
	spec := vm.InstanceSpec{
		Az:           GetStringAddr(d, "az"),
		InstanceType: GetStringAddr(d, "instance_type"),
		ImageId:      GetStringAddr(d, "image_id"),
		Name:         d.Get("instance_name").(string),
		PrimaryNetworkInterface: &vm.InstanceNetworkInterfaceAttachmentSpec{
			NetworkInterface: &vpc.NetworkInterfaceSpec{SubnetId: d.Get("subnet_id").(string)},
			//NetworkInterface: &vpc.NetworkInterfaceSpec{SubnetId: d.Get("subnet_id").(string), Az: GetStringAddr(d, "az")},
		},
	}

	if _, ok := d.GetOk("system_disk"); ok {
		spec.SystemDisk = &(typeListToDiskList(d.Get("system_disk").([]interface{}))[0])
	}

	if _, ok := d.GetOk("data_disk"); ok {
		d := typeListToDiskList(d.Get("data_disk").([]interface{}))
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

	if _, ok := d.GetOk("secondary_ip_count"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.SecondaryIpCount = GetIntAddr(d, "secondary_ip_count")
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

	// Just send a request here
	instanceId := ""
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.CreateInstances(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			instanceId = resp.Result.InstanceIds[0]
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

	// Waiting until VMs are ready
	err = instanceStatusWaiter(d, m, instanceId, []string{VM_PENDING, VM_STARTING}, []string{VM_RUNNING})
	if err != nil {
		return err
	}

	d.SetId(instanceId)
	return resourceJDCloudInstanceRead(d, m)
}

func resourceJDCloudInstanceRead(d *schema.ResourceData, m interface{}) error {

	vmInstanceDetail, err := QueryInstanceDetail(d, m, d.Id())

	if err != nil {
		return fmt.Errorf("[E] Failed in InstanceRead/QueryInstance %v", err)
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
	d.Set("az", vmInstanceDetail.Result.Instance.Az)
	d.Set("key_names", vmInstanceDetail.Result.Instance.KeyNames)

	if errSet := d.Set("security_group_ids", vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.SecurityGroups); err != nil {
		return fmt.Errorf("[ERROR] Failed in setting Sg Id LIST, reasons:%s", errSet.Error())
	}

	if errSet := d.Set("ip_addresses", ipList(vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.SecondaryIps)); err != nil {
		return fmt.Errorf("[ERROR] Failed in setting secondary ip LIST, reasons:%s", errSet.Error())
	}

	if errSet := d.Set("data_disk", cloudDiskStructIntoMap(vmInstanceDetail.Result.Instance.DataDisks)); err != nil {
		return fmt.Errorf("[ERROR] Failed in setting data_disk, reasons:%s", errSet.Error())
	}

	if errSet := d.Set("system_disk", sysDiskStructIntoMap(vmInstanceDetail.Result.Instance.SystemDisk)); err != nil {
		return fmt.Errorf("[ERROR] Failed in setting system_disk, reasons:%s", errSet.Error())
	}

	return nil
}

func resourceJDCloudInstanceUpdate(d *schema.ResourceData, m interface{}) error {

	d.Partial(true)
	defer d.Partial(false)
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)

	if d.HasChange("instance_name") || d.HasChange("description") {

		req := apis.NewModifyInstanceAttributeRequestWithAllParams(config.Region, d.Id(), GetStringAddr(d, "instance_name"), GetStringAddr(d, "description"))
		err := resource.Retry(time.Minute, func() *resource.RetryError {
			_, e := vmClient.ModifyInstanceAttribute(req)
			if connectionError(e) {
				return resource.RetryableError(e)
			} else {
				return resource.NonRetryableError(e)
			}
		})
		if err != nil {
			return err
		}

		d.SetPartial("instance_name")
		d.SetPartial("description")
	}

	if d.HasChange("password") {
		// Stop VM
		if err := StopVmInstance(d, m, d.Id()); err != nil {
			return fmt.Errorf("stop instance got error:%s", err)
		}
		if err := instanceStatusWaiter(d, m, d.Id(), []string{VM_RUNNING, VM_STOPPING}, []string{VM_STOPPED, VM_STOPPED_2}); err != nil {
			return fmt.Errorf("stop instance got error(2):%s", err)
		}

		//  Modify password
		req := apis.NewModifyInstancePasswordRequest(config.Region, d.Id(), d.Get("password").(string))
		err := resource.Retry(time.Minute, func() *resource.RetryError {
			_, e := vmClient.ModifyInstancePassword(req)
			if connectionError(e) {
				return resource.RetryableError(e)
			} else {
				return resource.NonRetryableError(e)
			}
		})
		if err != nil {
			return err
		}

		// Then start it
		if err := StartVmInstance(d, m); err != nil {
			return fmt.Errorf("start instance got error:%s", err)
		}
		d.SetPartial("password")
	}

	return resourceJDCloudInstanceRead(d, m)
}

func resourceJDCloudInstanceDelete(d *schema.ResourceData, m interface{}) error {

	// Stop VM
	err := StopVmInstance(d, m, d.Id())
	if err != nil {
		return fmt.Errorf("stop instance got error:%s", err)
	}

	// Wait until stopped
	err = instanceStatusWaiter(d, m, d.Id(), []string{VM_RUNNING, VM_STOPPING}, []string{VM_STOPPED, VM_STOPPED_2})
	if err != nil {
		return err
	}

	// Delete VM
	err = DeleteVmInstance(d, m, d.Id())
	if err != nil {
		return err
	}

	// Wait until deleted
	err = instanceStatusWaiter(d, m, d.Id(), []string{VM_RUNNING, VM_STOPPING, VM_DELETING}, []string{VM_DELETED})
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
