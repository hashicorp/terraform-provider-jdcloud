package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"time"
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
	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := vmClient.AssociateElasticIp(rq)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.RequestID)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
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

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DisassociateElasticIp(rq)

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
}
