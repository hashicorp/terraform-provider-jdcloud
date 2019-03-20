package jdcloud

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vpcApis "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	vpcClient "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
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
	err := resource.Retry(time.Minute, func() *resource.RetryError {

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

	if err != nil {
		return err
	}

	return resourceAssociateElasticIpRead(d, meta)
}
func resourceAssociateElasticIpRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	elasticIpId := d.Get("elastic_ip_id").(string)
	c := vpcClient.NewVpcClient(config.Credential)
	req := vpcApis.NewDescribeElasticIpRequest(config.Region, elasticIpId)

	return resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err := c.DescribeElasticIp(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.Set("elastic_ip_id", resp.Result.ElasticIp.ElasticIpId)
			d.Set("instance_id", resp.Result.ElasticIp.InstanceId)
			return nil
		}

		if resp.Result.ElasticIp.InstanceId != instanceID {
			log.Printf("[WARN] EIP=%s removed locally", resp.Result.ElasticIp.ElasticIpAddress)
			d.SetId("")
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
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
