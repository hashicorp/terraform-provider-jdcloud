package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"log"
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

	//构造请求
	rq := apis.NewAssociateElasticIpRequest(config.Region, instanceID, elasticIpId)

	//发送请求
	resp, err := vmClient.AssociateElasticIp(rq)

	if err != nil {
		log.Printf("[DEBUG] resourceAssociateElasticIpCreate failed %s ", err.Error())
		return err
	} else if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceAssociateElasticIpCreate  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.RequestID)

	return nil
}
func resourceAssociateElasticIpRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceAssociateElasticIpDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	instanceID := d.Get("instance_id").(string)
	elasticIpId := d.Get("elastic_ip_id").(string)

	vmClient := client.NewVmClient(config.Credential)

	//构造请求
	rq := apis.NewDisassociateElasticIpRequest(config.Region, instanceID, elasticIpId)

	//发送请求
	resp, err := vmClient.DisassociateElasticIp(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceAssociateElasticIpDelete failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceAssociateElasticIpDelete  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	//TODO 查询确认卸载

	return nil
}
