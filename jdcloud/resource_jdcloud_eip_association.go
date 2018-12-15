package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
)

func resourceJDCloudAssociateElasticIp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAssociateElasticIpCreate,
		Read:   resourceAssociateElasticIpRead,
		Delete: resourceAssociateElasticIpDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringNoEmpty,
			},
			"elastic_ip_id": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateStringNoEmpty,
			},
		},
	}
}

func resourceAssociateElasticIpCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	elasticIpId := d.Get("elastic_ip_id").(string)

	vmClient := client.NewVmClient(config.Credential)
	rq := apis.NewAssociateElasticIpRequest(config.Region, instanceID, elasticIpId)
	resp, err := vmClient.AssociateElasticIp(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceAssociateElasticIpCreate failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceAssociateElasticIpCreate code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.RequestID)

	return nil
}
func resourceAssociateElasticIpRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	elasticIpId := d.Get("elastic_ip_id").(string)

	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstanceRequest(config.Region, instanceID)
	resp, err := vmClient.DescribeInstance(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceAssociateElasticIpRead failed %s ", err.Error())
	}
	if resp.Result.Instance.ElasticIpId != elasticIpId {
		d.SetId("")
	}

	return nil
}

func resourceAssociateElasticIpDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	elasticIpId := d.Get("elastic_ip_id").(string)
	rq := apis.NewDisassociateElasticIpRequest(config.Region, instanceID, elasticIpId)

	vmClient := client.NewVmClient(config.Credential)
	resp, err := vmClient.DisassociateElasticIp(rq)

	if err != nil {
		return fmt.Errorf("[DEBUG] resourceAssociateElasticIpDelete failed %s ", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[DEBUG] resourceAssociateElasticIpDelete  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}
