package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
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
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudNetworkSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcId := d.Get("vpc_id").(string)
	networkSecurityGroupName := d.Get("network_security_group_name").(string)

	vpcClient := client.NewVpcClient(config.Credential)
	rq := apis.NewCreateNetworkSecurityGroupRequest(config.Region, vpcId, networkSecurityGroupName)
	if descriptionInterface, ok := d.GetOk("description"); ok {
		description := descriptionInterface.(string)
		rq.Description = &description
	}

	resp, err := vpcClient.CreateNetworkSecurityGroup(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupCreate failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.Result.NetworkSecurityGroupId)
	return nil
}

func resourceJDCloudNetworkSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {

	config   := meta.(*JDCloudConfig)
	sgClient := client.NewVpcClient(config.Credential)

	regionId := config.Region
	sgId     := d.Id()

	req 	 := apis.NewDescribeNetworkSecurityGroupRequest(regionId,sgId)
	resp,err := sgClient.DescribeNetworkSecurityGroup(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRead failed %s ", err.Error())
	}

	if resp.Error.Code == 404 {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.Set("description",resp.Result.NetworkSecurityGroup.Description)
	d.Set("network_security_group_name",resp.Result.NetworkSecurityGroup.NetworkSecurityGroupName)
	d.Set("vpc_id",resp.Result.NetworkSecurityGroup.VpcId)
	return nil
}

func resourceJDCloudNetworkSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {

	config   := meta.(*JDCloudConfig)
	sgClient := client.NewVpcClient(config.Credential)

	if d.HasChange("network_security_group_name") || d.HasChange("description"){
		req := apis.NewModifyNetworkSecurityGroupRequestWithAllParams(config.Region,d.Id(),GetStringAddr(d,"description"),GetStringAddr(d,"network_security_group_name"))
		resp,err := sgClient.ModifyNetworkSecurityGroup(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupUpdate failed %s ", err.Error())
		}
		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}
	}
	return nil
}


func resourceJDCloudNetworkSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	networkSecurityGroupId := d.Id()
	vpcClient := client.NewVpcClient(config.Credential)

	rq := apis.NewDeleteNetworkSecurityGroupRequest(config.Region, networkSecurityGroupId)
	resp, err := vpcClient.DeleteNetworkSecurityGroup(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupDelete failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkSecurityGroupDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.SetId("")
	return nil
}
