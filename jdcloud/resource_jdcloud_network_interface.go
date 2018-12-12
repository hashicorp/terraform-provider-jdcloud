package jdcloud

import (
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
				Type: schema.TypeList,
				// Optional : Can be provided by user
				// Computed : Can be provided by computed
				Optional: true,
				Computed: true,
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

	req := apis.NewCreateNetworkInterfaceRequest(config.Region, subnetID)
	networkInterfaceName := d.Get("network_interface_name").(string)
	req.NetworkInterfaceName = &networkInterfaceName

	if _, ok := d.GetOk("az"); ok {
		req.Az = GetStringAddr(d, "az")
	}

	if _, ok := d.GetOk("description"); ok {
		req.Description = GetStringAddr(d, "description")
	}

	if _, ok := d.GetOk("sanity_check"); ok {
		req.SanityCheck = GetIntAddr(d, "sanity_check")
	}

	if v, ok := d.GetOk("secondary_ip_addresses"); ok {
		for _, vv := range v.([]interface{}) {
			secondaryIpAddress := vv.(string)
			req.SecondaryIpAddresses = append(req.SecondaryIpAddresses, secondaryIpAddress)
		}
	}

	if secondaryIpCountInterface, ok := d.GetOk("secondary_ip_count"); ok {
		secondaryIpCount := secondaryIpCountInterface.(int)
		req.SecondaryIpCount = &secondaryIpCount
	}

	setDefaultSecurityGroup := true
	if sgArray, ok := d.GetOk("security_groups"); ok {
		setDefaultSecurityGroup = false
		for _, sg := range sgArray.([]interface{}) {
			req.SecurityGroups = append(req.SecurityGroups, sg.(string))
		}
	}

	vpcClient := client.NewVpcClient(config.Credential)
	resp, err := vpcClient.CreateNetworkInterface(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceCreate failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.Result.NetworkInterfaceId)
	d.Set("network_interface_id", resp.Result.NetworkInterfaceId)

	// Default sgID is set and retrieved via "READ"
	if setDefaultSecurityGroup {
		errNIRead := resourceJDCloudNetworkInterfaceRead(d, meta)
		if errNIRead != nil {
			log.Printf("[WARN] NI has been created but failed to update info, commmand 'Terraform refresh'")
			log.Printf("[WARN] to update your local info again'")
		}
	}

	return nil
}

func resourceJDCloudNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	networkInterfaceClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeNetworkInterfaceRequest(config.Region, d.Id())
	resp, err := networkInterfaceClient.DescribeNetworkInterface(req)

	if err != nil {
		return err
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceRead failed error code:%d, message:%s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result.NetworkInterface.Az != "" {
		d.Set("az", resp.Result.NetworkInterface.Az)
	}

	if resp.Result.NetworkInterface.Description != "" {
		d.Set("description", resp.Result.NetworkInterface.Description)
	}

	if resp.Result.NetworkInterface.NetworkInterfaceName != "" {
		d.Set("network_interface_name", resp.Result.NetworkInterface.NetworkInterfaceName)
	}

	if resp.Result.NetworkInterface.SanityCheck != 0 {
		d.Set("sanity_check", resp.Result.NetworkInterface.SanityCheck)
	}

	if resp.Result.NetworkInterface.PrimaryIp.ElasticIpAddress != "" {
		d.Set("primary_ip_address", resp.Result.NetworkInterface.PrimaryIp.ElasticIpAddress)
	}

	if len(resp.Result.NetworkInterface.SecondaryIps) != 0 {
		d.Set("secondary_ip_addresses", resp.Result.NetworkInterface.SecondaryIps)
	}

	sgRemote := resp.Result.NetworkInterface.NetworkSecurityGroupIds
	sgLocal := InterfaceToStringArray(d.Get("security_groups").([]interface{}))

	if len(sgRemote) != 0 && equalSliceString(sgRemote, sgLocal) == false {
		d.Set("security_groups", resp.Result.NetworkInterface.NetworkSecurityGroupIds)
	}

	return nil
}

func resourceJDCloudNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("network_interface_name") || d.HasChange("secondary_ip_addresses") || d.HasChange("security_groups") {

		config := meta.(*JDCloudConfig)
		vpcClient := client.NewVpcClient(config.Credential)

		sg := InterfaceToStringArray(d.Get("security_groups").([]interface{}))
		req := apis.NewModifyNetworkInterfaceRequestWithAllParams(config.Region, d.Id(), GetStringAddr(d, "network_interface_name"), GetStringAddr(d, "description"), sg)
		resp, err := vpcClient.ModifyNetworkInterface(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceUpdate failed %s ", err.Error())
		}

		if resp.Error.Code != 0 {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

	}

	return nil
}

func resourceJDCloudNetworkInterfaceDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	networkInterfaceId := d.Get("network_interface_id").(string)
	rq := apis.NewDeleteNetworkInterfaceRequest(config.Region, networkInterfaceId)

	vpcClient := client.NewVpcClient(config.Credential)
	resp, err := vpcClient.DeleteNetworkInterface(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceDelete failed %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.SetId("")
	return nil
}
