package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpcModels "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
)

func resourceJDCloudEIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudEIPCreate,
		Read:   resourceJDCloudEIPRead,
		Update: resourceJDCloudEIPUpdate,
		Delete: resourceJDCloudEIPDelete,

		Schema: map[string]*schema.Schema{
			"eip_provider": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"bandwidth_mbps": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"elastic_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceJDCloudEIPCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	bandwidthMbps := d.Get("bandwidth_mbps").(int)
	provider := d.Get("eip_provider").(string)

	vpcClient := client.NewVpcClient(config.Credential)

	var elasticIpSpec vpcModels.ElasticIpSpec

	var ChargeSpec models.ChargeSpec

	elasticIpSpec.BandwidthMbps = bandwidthMbps
	elasticIpSpec.Provider = provider
	elasticIpSpec.ChargeSpec = &ChargeSpec

	//构造请求
	rq := apis.NewCreateElasticIpsRequest(config.Region, 1, &elasticIpSpec)

	if elasticIpAddressInterface, ok := d.GetOk("elastic_ip_address"); ok {
		elasticIpAddress := elasticIpAddressInterface.(string)
		rq.ElasticIpAddress = &elasticIpAddress
	}

	//发送请求
	resp, err := vpcClient.CreateElasticIps(rq)

	if err != nil {
		log.Printf("[DEBUG] resourceJDCloudEIPCreate failed %s ", err.Error())
		return err
	} else if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudEIPCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.Result.ElasticIpIds[0])

	return nil
}

func resourceJDCloudEIPRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceJDCloudEIPUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceJDCloudEIPDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	//构造请求
	elasticIpId := d.Id()
	rq := apis.NewDeleteElasticIpRequest(config.Region, elasticIpId)

	//发送请求
	resp, err := vpcClient.DeleteElasticIp(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudEIPDelete failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudEIPDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	return nil
}
