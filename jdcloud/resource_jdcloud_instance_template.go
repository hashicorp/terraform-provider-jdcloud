package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	"time"
)

func resourceJDCloudInstanceTemplate() *schema.Resource {

	diskSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{

			// Disk-Attachment-Spec
			"auto_delete": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"device_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// Disk-Spec
			"disk_category": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringCandidates("local", "cloud"),
			},
			"disk_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "ssd",
				ValidateFunc: validateStringCandidates("ssd", "premium-hdd"),
			},
			"disk_size": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      40,
				ValidateFunc: validateDiskSize(),
			},
			"snapshot_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}

	return &schema.Resource{
		Create: resourceJDCloudInstanceTemplateCreate,
		Read:   resourceJDCloudInstanceTemplateRead,
		Update: resourceJDCloudInstanceTemplateUpdate,
		Delete: resourceJDCloudInstanceTemplateDelete,

		Schema: map[string]*schema.Schema{
			"template_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"instance_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"key_names"},
			},
			"key_names": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"password"},
			},
			"bandwidth": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ip_service_provider": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"charge_mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"security_group_ids": &schema.Schema{
				Type:     schema.TypeSet,
				MinItems: 1,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"system_disk": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     diskSchema,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				MinItems: 1,
			},
			"data_disks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     diskSchema,
			},
		},
	}
}

func resourceJDCloudInstanceTemplateCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	logger := vmLogger{}
	vmClient.SetLogger(logger)

	templateSpec := &vm.InstanceTemplateSpec{
		InstanceType: d.Get("instance_type").(string),
		ImageId:      d.Get("image_id").(string),
		KeyNames:     []string{},
		PrimaryNetworkInterface: vm.InstanceTemplateNetworkInterfaceAttachmentSpec{
			DeviceIndex: DEFAULT_DEVICE_INDEX,
			AutoDelete:  DEFAULT_NETWORK_INTERFACE_AUTO_DELETE,
			NetworkInterface: vm.InstanceTemplateNetworkInterfaceSpec{
				SubnetId:       d.Get("subnet_id").(string),
				SecurityGroups: typeSetToStringArray(d.Get("security_group_ids").(*schema.Set)),
				SanityCheck:    DEFAULT_SANITY_CHECK,
			},
		},
	}

	if _, ok := d.GetOk("bandwidth"); ok {
		templateSpec.ElasticIp = &vm.InstanceTemplateElasticIpSpec{
			BandwidthMbps: d.Get("bandwidth").(int),
			Provider:      d.Get("ip_service_provider").(string),
			ChargeMode:    d.Get("charge_mode").(string),
		}
	}
	if _, ok := d.GetOk("password"); ok {
		templateSpec.Password = d.Get("password").(string)
	}
	if _, ok := d.GetOk("system_disk"); ok {
		templateSpec.SystemDisk = typeListToDiskTemplateList(d.Get("system_disk").([]interface{}))[0]
	}
	if _, ok := d.GetOk("key_names"); ok {
		templateSpec.KeyNames = []string{d.Get("key_names").(string)}
	}
	if _, ok := d.GetOk("data_disks"); ok {
		templateSpec.DataDisks = typeListToDiskTemplateList(d.Get("data_disks").([]interface{}))
	}
	req := apis.NewCreateInstanceTemplateRequest(config.Region, templateSpec, d.Get("template_name").(string))
	if _, ok := d.GetOk("description"); ok {
		req.Description = GetStringAddr(d, "description")
	}

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.CreateInstanceTemplate(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.InstanceTemplateId)
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
	return resourceJDCloudInstanceTemplateRead(d, m)
}

func resourceJDCloudInstanceTemplateRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceTemplateRequest(config.Region, d.Id())
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DescribeInstanceTemplate(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {

			d.Set("subnet_id", resp.Result.InstanceTemplate.InstanceTemplateData.PrimaryNetworkInterface.NetworkInterface.SubnetId)
			d.Set("ip_service_provider", resp.Result.InstanceTemplate.InstanceTemplateData.ElasticIp.Provider)
			d.Set("charge_mode", resp.Result.InstanceTemplate.InstanceTemplateData.ElasticIp.ChargeMode)
			d.Set("bandwidth", resp.Result.InstanceTemplate.InstanceTemplateData.ElasticIp.BandwidthMbps)
			d.Set("template_name", resp.Result.InstanceTemplate.Name)
			d.Set("description", resp.Result.InstanceTemplate.Description)
			d.Set("image_id", resp.Result.InstanceTemplate.InstanceTemplateData.ImageId)
			d.Set("instance_type", resp.Result.InstanceTemplate.InstanceTemplateData.InstanceType)
			d.Set("template_name", resp.Result.InstanceTemplate.Name)

			if e := d.Set("data_disks", typeListToDiskTemplateMap(resp.Result.InstanceTemplate.InstanceTemplateData.DataDisks)); e != nil {
				return resource.NonRetryableError(fmt.Errorf("[E] Failed in setting data disks"))
			}

			if len(resp.Result.InstanceTemplate.InstanceTemplateData.KeyNames) > 0 {
				d.Set("key_names", resp.Result.InstanceTemplate.InstanceTemplateData.KeyNames[0])
			}
			sysDisk := typeListToDiskTemplateMap([]vm.InstanceTemplateDiskAttachment{resp.Result.InstanceTemplate.InstanceTemplateData.SystemDisk})
			sysDisk[0]["disk_type"] = d.Get("system_disk.0.disk_type")
			sysDisk[0]["device_name"] = d.Get("system_disk.0.device_name")
			if e := d.Set("system_disk", sysDisk); e != nil {
				return resource.NonRetryableError(fmt.Errorf("[E] Failed in setting data disks"))
			}
			if e := d.Set("security_group_ids", resp.Result.InstanceTemplate.InstanceTemplateData.PrimaryNetworkInterface.NetworkInterface.SecurityGroups); e != nil {
				return resource.NonRetryableError(fmt.Errorf("[E] Failed in setting data disks"))
			}

			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND {
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

func resourceJDCloudInstanceTemplateUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("template_name") {
		config := m.(*JDCloudConfig)
		vmClient := client.NewVmClient(config.Credential)
		req := apis.NewUpdateInstanceTemplateRequestWithAllParams(config.Region, d.Id(), nil, stringAddr(d.Get("template_name")))

		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.UpdateInstanceTemplate(req)
			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				return nil
			}

			if resp.Error.Code == RESOURCE_NOT_FOUND {
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
	}
	return resourceJDCloudInstanceTemplateRead(d, m)
}

func resourceJDCloudInstanceTemplateDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDeleteInstanceTemplateRequest(config.Region, d.Id())

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DeleteInstanceTemplate(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
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

func typeListToDiskTemplateMap(s []vm.InstanceTemplateDiskAttachment) []map[string]interface{} {

	ret := []map[string]interface{}{}
	for _, item := range s {
		ret = append(ret, map[string]interface{}{
			"disk_size":     item.InstanceTemplateDisk.DiskSizeGB,
			"disk_type":     item.InstanceTemplateDisk.DiskType,
			"disk_category": item.DiskCategory,
			"device_name":   item.DeviceName,
			"auto_delete":   item.AutoDelete,
		})
	}
	return ret
}
func typeListToDiskTemplateList(s []interface{}) []vm.InstanceTemplateDiskAttachmentSpec {

	ret := []vm.InstanceTemplateDiskAttachmentSpec{}

	for _, d := range s {
		m := d.(map[string]interface{})
		disk := vm.InstanceTemplateDiskAttachmentSpec{
			DiskCategory: m["disk_category"].(string),
			CloudDiskSpec: vm.InstanceTemplateDiskSpec{
				DiskType:   m["disk_type"].(string),
				DiskSizeGB: m["disk_size"].(int),
			},
		}

		if m["snapshot_id"] != "" {
			disk.CloudDiskSpec.SnapshotId = m["snapshot_id"].(string)
		}
		if m["auto_delete"] != "" {
			disk.AutoDelete = m["auto_delete"].(bool)
		}
		if m["device_name"] != "" {
			disk.DeviceName = m["device_name"].(string)
		}
		ret = append(ret, disk)
	}

	return ret
}
