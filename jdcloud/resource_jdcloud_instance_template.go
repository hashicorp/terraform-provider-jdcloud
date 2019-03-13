package jdcloud

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	"strconv"
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
			},
			"device_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			// Disk-Spec
			"disk_category": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStringCandidates("local", "cloud"),
			},
			"disk_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ssd",
				ValidateFunc: validateStringCandidates("ssd", "premium-hdd"),
			},
			"disk_size": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      40,
				ValidateFunc: validateDiskSize(),
			},
			"snapshot_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
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
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"bandwidth": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"ip_service_provider": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "BGP",
			},
			"charge_mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "bandwith",
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
				Type:     schema.TypeSet,
				Elem:     diskSchema,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				MinItems: 1,
			},
			"data_disks": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     diskSchema,
				MinItems: 1,
			},
		},
	}
}

func resourceJDCloudInstanceTemplateCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)

	templateSpec := &vm.InstanceTemplateSpec{
		InstanceType: d.Get("instance_type").(string),
		ImageId:      d.Get("image_id").(string),
		Password:     d.Get("password").(string),
		KeyNames:     []string{},
		ElasticIp: vm.InstanceTemplateElasticIpSpec{
			BandwidthMbps: d.Get("bandwidth").(int),
			Provider:      d.Get("ip_service_provider").(string),
			ChargeMode:    d.Get("charge_mode").(string),
		},
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

	if _, ok := d.GetOk("system_disk"); ok {
		templateSpec.SystemDisk = typeSetToDiskTemplateList(d.Get("system_disk").(*schema.Set))[0]
	}
	if _, ok := d.GetOk("data_disks"); ok {
		templateSpec.DataDisks = typeSetToDiskTemplateList(d.Get("data_disks").(*schema.Set))
	}

	req := apis.NewCreateInstanceTemplateRequest(config.Region, templateSpec, d.Get("template_name").(string))

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
	return nil
}

func resourceJDCloudInstanceTemplateRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceTemplateRequest(config.Region, d.Id())
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DescribeInstanceTemplate(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.Set("template_name", resp.Result.InstanceTemplate.Name)
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
	return nil
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

func stringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func typeSetToDiskTemplateList(s *schema.Set) []vm.InstanceTemplateDiskAttachmentSpec {

	ret := []vm.InstanceTemplateDiskAttachmentSpec{}

	for _, d := range s.List() {
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
