package jdcloud

import (
	"errors"
	"fmt"
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

	return nil

}

func resourceJDCloudNetworkSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {

	config   := meta.(*JDCloudConfig)
	sgClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	sgId     := d.Get("network_security_group_id").(string)

	req 	 := apis.NewDescribeNetworkSecurityGroupRequest(regionId,sgId)
	resp,err := sgClient.DescribeNetworkSecurityGroup(req)

	if err!=nil {
		return err
	}
	if resp.Error.Code!=0 {
		return fmt.Errorf("failed in creating new security group, fail info shown as below:%s",resp.Error)
	}

	d.Set("description",resp.Result.NetworkSecurityGroup.Description)
	d.Set("network_security_group_name",resp.Result.NetworkSecurityGroup.NetworkSecurityGroupName)
	d.Set("vpc_id",resp.Result.NetworkSecurityGroup.VpcId)

	return nil
}

func resourceJDCloudNetworkSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}
func resourceJDCloudNetworkSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	networkSecurityGroupId := d.Id()

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
