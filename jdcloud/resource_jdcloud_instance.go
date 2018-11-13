package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"time"
)

func resourceJDCloudInstance() *schema.Resource {
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
			"disk_category": {
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

			"private_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"security_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"public_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func GetStringAddr(d *schema.ResourceData, key string) *string {
	v := d.Get(key).(string)
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
	if resp.Error.Code == 404 && resp.Error.Status == "NOT_FOUND" {
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

func resourceJDCloudInstanceCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)

	spec := vm.InstanceSpec{
		Az:           GetStringAddr(d, "az"),
		InstanceType: GetStringAddr(d, "instance_type"),
		ImageId:      GetStringAddr(d, "image_id"),
		Name:         d.Get("instance_name").(string),
		PrimaryNetworkInterface: &vm.InstanceNetworkInterfaceAttachmentSpec{
			NetworkInterface: &vpc.NetworkInterfaceSpec{SubnetId: d.Get("subnet_id").(string), Az: GetStringAddr(d, "az")},
		},
		SystemDisk: &vm.InstanceDiskAttachmentSpec{DiskCategory: GetStringAddr(d, "disk_category")},
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

	if _, ok := d.GetOk("private_id"); ok {
		spec.PrimaryNetworkInterface.NetworkInterface.PrimaryIpAddress = GetStringAddr(d, "private_id")
	}

	if sgs, ok := d.GetOk("security_groups"); ok {
		sgList := InterfaceToStringArray(sgs.(*schema.Set).List())
		if len(sgList) > DefaultSecurityGroupsMax {
			return fmt.Errorf("the maximum allowed number of security_groups is %d", DefaultSecurityGroupsMax)
		}
		spec.PrimaryNetworkInterface.NetworkInterface.SecurityGroups = sgList
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
	d.Set("private_id", vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.PrimaryIp)
	d.Set("key_names", vmInstanceDetail.Result.Instance.KeyNames)
	d.Set("security_groups", vmInstanceDetail.Result.Instance.PrimaryNetworkInterface.NetworkInterface.SecurityGroups)
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
