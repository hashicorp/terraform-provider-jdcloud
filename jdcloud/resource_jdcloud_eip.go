package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/charge/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	vpcModels "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
)

// Only one EIP is allowed to create in each resource
const maxEIPCount = 1

func resourceJDCloudEIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudEIPCreate,
		Read:   resourceJDCloudEIPRead,
		Delete: resourceJDCloudEIPDelete,

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
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudEIPCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	elasticIpSpec := vpcModels.ElasticIpSpec{
		d.Get("bandwidth_mbps").(int),
		d.Get("eip_provider").(string),
		&models.ChargeSpec{},
	}
	req := apis.NewCreateElasticIpsRequest(config.Region, maxEIPCount, &elasticIpSpec)
	if _, ok := d.GetOk("elastic_ip_address"); ok {
		req.ElasticIpAddress = GetStringAddr(d,"elastic_ip_address")
	}

	vpcClient := client.NewVpcClient(config.Credential)
	resp, err := vpcClient.CreateElasticIps(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudEIPCreate failed %s ", err.Error())
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudEIPCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.Result.ElasticIpIds[0])
	return nil
}

func resourceJDCloudEIPRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	req := apis.NewDescribeElasticIpRequest(config.Region,d.Id())
	vpcClient := client.NewVpcClient(config.Credential)
	resp,err := vpcClient.DescribeElasticIp(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudEIPRead failed %s ", err.Error())
	}

	if resp.Error.Code == 404 {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudEIPRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}

//  TODO EIP deletion may take some time
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
