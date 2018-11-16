package jdcloud

import (
	"errors"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
)

func resourceJDCloudNetworkSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudNetworkSecurityGroupCreate,
		Read:   resourceJDCloudNetworkSecurityGroupRead,
		Update: resourceJDCloudNetworkSecurityGroupUpdate,
		Delete: resourceJDCloudNetworkSecurityGroupDelete,

		Schema: map[string]*schema.Schema{
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_security_group_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"network_security_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceJDCloudNetworkSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	vpcId := d.Get("vpc_id").(string)
	networkSecurityGroupName := d.Get("network_security_group_name").(string)

	vpcClient := client.NewVpcClient(config.Credential)

	//构造请求
	rq := apis.NewCreateNetworkSecurityGroupRequest(config.Region, vpcId, networkSecurityGroupName)

	if descriptionInterface, ok := d.GetOk("description"); ok {
		description := descriptionInterface.(string)
		rq.Description = &description
	}

	//发送请求
	resp, err := vpcClient.CreateNetworkSecurityGroup(rq)

	if err != nil {

		log.Printf("[DEBUG] CreateNetworkSecurityGroup failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] CreateNetworkSecurityGroup failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.Result.NetworkSecurityGroupId)
	d.Set("network_security_group_id", resp.Result.NetworkSecurityGroupId)

	return nil

}

func resourceJDCloudNetworkSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceJDCloudNetworkSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}
func resourceJDCloudNetworkSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	networkSecurityGroupId := d.Get("network_security_group_id").(string)

	vpcClient := client.NewVpcClient(config.Credential)

	//构造请求
	rq := apis.NewDeleteNetworkSecurityGroupRequest(config.Region, networkSecurityGroupId)

	//发送请求
	resp, err := vpcClient.DeleteNetworkSecurityGroup(rq)

	if err != nil {

		log.Printf("[DEBUG] CreateNetworkSecurityGroup failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] CreateNetworkSecurityGroup failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	return nil
}
