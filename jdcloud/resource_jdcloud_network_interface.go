package jdcloud

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
)

func resourceJDCloudNetworkInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudNetworkInterfaceCreate,
		Read:   resourceJDCloudNetworkInterfaceRead,
		Update: resourceJDCloudNetworkInterfaceUpdate,
		Delete: resourceJDCloudNetworkInterfaceDelete,

		Schema: map[string]*schema.Schema{
			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"network_interface_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"az": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"primary_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"sanity_check": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"secondary_ip_addresses": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"secondary_ip_count": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"security_groups": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MaxItems: DefaultSecurityGroupsMax,
			},
			"network_interface_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceJDCloudNetworkInterfaceCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	subnetID := d.Get("subnet_id").(string)

	vpcClient := client.NewVpcClient(config.Credential)

	rq := apis.NewCreateNetworkInterfaceRequest(config.Region, subnetID)

	networkInterfaceName := d.Get("network_interface_name").(string)

	rq.NetworkInterfaceName = &networkInterfaceName

	if availabilityZoneInterface, ok := d.GetOk("az"); ok {
		availabilityZone := availabilityZoneInterface.(string)
		rq.Az = &availabilityZone
	}

	if descriptionInterface, ok := d.GetOk("description"); ok {
		description := descriptionInterface.(string)
		rq.Description = &description
	}

	if sanityCheckInterface, ok := d.GetOk("sanity_check"); ok {
		sanityCheck := sanityCheckInterface.(int)
		rq.SanityCheck = &sanityCheck
	}

	if v, ok := d.GetOk("secondary_ip_addresses"); ok {

		for _, vv := range v.([]interface{}) {

			secondaryIpAddress := vv.(string)
			rq.SecondaryIpAddresses = append(rq.SecondaryIpAddresses, secondaryIpAddress)
		}
	}
	// This parameter is only used in creating network interface
	// Never update later
	if secondaryIpCountInterface, ok := d.GetOk("secondary_ip_count"); ok {
		secondaryIpCount := secondaryIpCountInterface.(int)
		rq.SecondaryIpCount = &secondaryIpCount
	}

	if v, ok := d.GetOk("security_groups"); ok {

		for _, vv := range v.([]interface{}) {

			securityGroup := vv.(string)
			rq.SecurityGroups = append(rq.SecurityGroups, securityGroup)
		}
	}

	resp, err := vpcClient.CreateNetworkInterface(rq)

	if err != nil {
		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceCreate failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}

	d.SetId(resp.Result.NetworkInterfaceId)
	d.Set("network_interface_id", resp.Result.NetworkInterfaceId)

	return nil

}

func resourceJDCloudNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {

	networkInterfaceConfig := meta.(*JDCloudConfig)
	networkInterfaceClient := client.NewVpcClient(networkInterfaceConfig.Credential)

	requestOnNetworkInterface := apis.NewDescribeNetworkInterfaceRequest(networkInterfaceConfig.Region,d.Id())
	responseOnNetworkInterface, err := networkInterfaceClient.DescribeNetworkInterface(requestOnNetworkInterface)

	if err != nil {
		return err
	}
	if responseOnNetworkInterface.Error.Code != 0{
		return fmt.Errorf("%s",responseOnNetworkInterface.Error)
	}

	if responseOnNetworkInterface.Result.NetworkInterface.Az != ""{
		d.Set("az",responseOnNetworkInterface.Result.NetworkInterface.Az)
	}
	if responseOnNetworkInterface.Result.NetworkInterface.Description != ""{
		d.Set("description",responseOnNetworkInterface.Result.NetworkInterface.Description)
	}
	if responseOnNetworkInterface.Result.NetworkInterface.NetworkInterfaceName != ""{
		d.Set("network_interface_name",responseOnNetworkInterface.Result.NetworkInterface.NetworkInterfaceName)
	}

	if responseOnNetworkInterface.Result.NetworkInterface.SanityCheck != 0{
		d.Set("sanity_check",responseOnNetworkInterface.Result.NetworkInterface.SanityCheck)
	}

	if len(responseOnNetworkInterface.Result.NetworkInterface.SecondaryIps) != 0{
		d.Set("secondary_ip_addresses",responseOnNetworkInterface.Result.NetworkInterface.SecondaryIps)
	}

	if len(responseOnNetworkInterface.Result.NetworkInterface.NetworkSecurityGroupIds) != 0 {
		d.Set("security_groups",responseOnNetworkInterface.Result.NetworkInterface.NetworkSecurityGroupIds)
	}

	// Not sure if we have to place entire struct into filed,
	// Currently just place PrimaryIp.ElasticIpAddress into PrimaryIp
	if responseOnNetworkInterface.Result.NetworkInterface.PrimaryIp.ElasticIpAddress != ""{
		d.Set("primary_ip_address",responseOnNetworkInterface.Result.NetworkInterface.PrimaryIp.ElasticIpAddress)
	}

	return nil
}
func resourceJDCloudNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {

	networkInterfaceConfig := meta.(*JDCloudConfig)
	networkInterfaceClient := client.NewVpcClient(networkInterfaceConfig.Credential)

	sg := InterfaceToStringArray(d.Get("security_groups").([]interface{}))
	requestOnNetworkInterface := apis.NewModifyNetworkInterfaceRequestWithAllParams(
					networkInterfaceConfig.Region, d.Id(),GetStringAddr(d,"network_interface_name"),
					GetStringAddr(d,"description"),sg)
	responseOnNetworkInterface, err := networkInterfaceClient.ModifyNetworkInterface(requestOnNetworkInterface)

	if err != nil{
		return err
	}
	if responseOnNetworkInterface.Error.Code!=0{
		return fmt.Errorf("%s",responseOnNetworkInterface.Error)
	}

	return nil
}

func resourceJDCloudNetworkInterfaceDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	networkInterfaceId := d.Get("network_interface_id").(string)

	vpcClient := client.NewVpcClient(config.Credential)

	//构造请求
	rq := apis.NewDeleteNetworkInterfaceRequest(config.Region, networkInterfaceId)

	//发送请求
	resp, err := vpcClient.DeleteNetworkInterface(rq)

	if err != nil {

		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceDelete failed %s ", err.Error())
		return err
	}

	if resp.Error.Code != 0 {
		log.Printf("[DEBUG] resourceJDCloudNetworkInterfaceDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		return errors.New(resp.Error.Message)
	}
	d.SetId("")
	return nil
}
