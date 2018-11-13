package jdcloud

import (
	"errors"
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
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"az": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_interface_name": &schema.Schema{
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

	if avalidZoneInterface, ok := d.GetOk("az"); ok {
		avalidZone := avalidZoneInterface.(string)
		rq.Az = &avalidZone
	}

	if descriptionInterface, ok := d.GetOk("description"); ok {
		description := descriptionInterface.(string)
		rq.Description = &description
	}

	if networkInterfaceNameInterface, ok := d.GetOk("network_interface_name"); ok {
		networkInterfaceName := networkInterfaceNameInterface.(string)
		rq.NetworkInterfaceName = &networkInterfaceName
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

	d.SetId(resp.RequestID)
	d.Set("network_interface_id", resp.Result.NetworkInterfaceId)

	return nil

}

func resourceJDCloudNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
func resourceJDCloudNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
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

	return nil
}
