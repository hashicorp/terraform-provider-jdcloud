package jdcloud

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpcModels "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"time"
)

func resourceJDCloudEIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudEIPCreate,
		Read:   resourceJDCloudEIPRead,
		Delete: resourceJDCloudEIPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"eip_provider": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bandwidth_mbps": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"elastic_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudEIPCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	elasticIpSpec := vpcModels.ElasticIpSpec{
		BandwidthMbps: d.Get("bandwidth_mbps").(int),
		Provider:      d.Get("eip_provider").(string),
		ChargeSpec:    &models.ChargeSpec{},
	}
	vpcClient := client.NewVpcClient(config.Credential)
	req := apis.NewCreateElasticIpsRequest(config.Region, MAX_EIP_COUNT, &elasticIpSpec)
	if _, ok := d.GetOk("elastic_ip_address"); ok {
		req.ElasticIpAddress = GetStringAddr(d, "elastic_ip_address")
	}

	err := resource.Retry(20*time.Second, func() *resource.RetryError {

		resp, err := vpcClient.CreateElasticIps(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.ElasticIpIds[0])
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

	return resourceJDCloudEIPRead(d, meta)
}

func resourceJDCloudEIPRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	req := apis.NewDescribeElasticIpRequest(config.Region, d.Id())
	vpcClient := client.NewVpcClient(config.Credential)

	return resource.Retry(time.Minute, func() *resource.RetryError {
		resp, err := vpcClient.DescribeElasticIp(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {

			d.Set("elastic_ip_address", resp.Result.ElasticIp.ElasticIpAddress)
			d.Set("bandwidth_mbps", resp.Result.ElasticIp.BandwidthMbps)
			d.Set("eip_provider", resp.Result.ElasticIp.Provider)

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
}

func resourceJDCloudEIPDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	elasticIpId := d.Id()
	rq := apis.NewDeleteElasticIpRequest(config.Region, elasticIpId)
	vpcClient := client.NewVpcClient(config.Credential)

	return resource.Retry(20*time.Second, func() *resource.RetryError {

		resp, err := vpcClient.DeleteElasticIp(rq)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId("")
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
}

/*
Log:
	Its been tested that you can not tell if an EIP has been detached or not
	Since there is no info on "status" about EIP
*/
