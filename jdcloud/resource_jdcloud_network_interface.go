package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/models"
	"log"
	"time"
)

func resourceJDCloudNetworkInterface() *schema.Resource {

	return &schema.Resource{
		Create: resourceJDCloudNetworkInterfaceCreate,
		Read:   resourceJDCloudNetworkInterfaceRead,
		Update: resourceJDCloudNetworkInterfaceUpdate,
		Delete: resourceJDCloudNetworkInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				ForceNew: true,
			},
			"primary_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"sanity_check": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"secondary_ip_addresses": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"secondary_ip_count": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"security_groups": &schema.Schema{
				Type: schema.TypeSet,
				// Optional : Can be provided by user
				// Computed : Can be provided by computed
				Optional: true,
				Computed: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MaxItems: MAX_SECURITY_GROUP_COUNT,
			},
			"network_interface_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceJDCloudNetworkInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)
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

	if _, ok := d.GetOk("secondary_ip_addresses"); ok {
		req.SecondaryIpAddresses = typeSetToStringArray(d.Get("secondary_ip_addresses").(*schema.Set))
	}

	if secondaryIpCountInterface, ok := d.GetOk("secondary_ip_count"); ok {
		secondaryIpCount := secondaryIpCountInterface.(int)
		req.SecondaryIpCount = &secondaryIpCount
	}

	sg := false
	if _, ok := d.GetOk("security_groups"); ok {
		req.SecurityGroups = typeSetToStringArray(d.Get("security_groups").(*schema.Set))
	}

	vpcClient := client.NewVpcClient(config.Credential)

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := vpcClient.CreateNetworkInterface(req)

		if err == nil {
			d.SetId(resp.Result.NetworkInterfaceId)
			d.Set("network_interface_id", resp.Result.NetworkInterfaceId)

			d.SetPartial("az")
			d.SetPartial("subnet_id")
			d.SetPartial("description")
			d.SetPartial("sanity_check")
			d.SetPartial("primary_ip_address")
			d.SetPartial("secondary_ip_count")
			d.SetPartial("network_interface_id")
			d.SetPartial("network_interface_name")
			d.SetPartial("secondary_ip_addresses")
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

	// Default sgID is set and retrieved via "READ"
	if sg {
		if errNIRead := resourceJDCloudNetworkInterfaceRead(d, meta); errNIRead != nil {
			return errNIRead
		}
		d.SetPartial("security_groups")
	}

	d.Partial(false)
	return nil
}

func resourceJDCloudNetworkInterfaceRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	networkInterfaceClient := client.NewVpcClient(config.Credential)

	req := apis.NewDescribeNetworkInterfaceRequest(config.Region, d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {

		resp, err := networkInterfaceClient.DescribeNetworkInterface(req)

		if err == nil {

			if resp.Result.NetworkInterface.Az != "" {
				d.Set("az", resp.Result.NetworkInterface.Az)
			}

			if resp.Result.NetworkInterface.Description != "" {
				d.Set("description", resp.Result.NetworkInterface.Description)
			}

			if resp.Result.NetworkInterface.NetworkInterfaceName != "" {
				d.Set("network_interface_name", resp.Result.NetworkInterface.NetworkInterfaceName)
			}

			if resp.Result.NetworkInterface.SanityCheck != REQUEST_COMPLETED {
				d.Set("sanity_check", resp.Result.NetworkInterface.SanityCheck)
			}

			if resp.Result.NetworkInterface.PrimaryIp.ElasticIpAddress != "" {
				d.Set("primary_ip_address", resp.Result.NetworkInterface.PrimaryIp.ElasticIpAddress)
			}
			if len(resp.Result.NetworkInterface.SecondaryIps) != d.Get("secondary_ip_count").(int)+d.Get("secondary_ip_addresses.#").(int) {
				if errSetIp := d.Set("secondary_ip_addresses", ipList(resp.Result.NetworkInterface.SecondaryIps)); errSetIp != nil {
					return resource.NonRetryableError(formatArraySetErrorMessage(errSetIp))
				}
			}

			// sg - Exclude default sg
			sgRemote := resp.Result.NetworkInterface.NetworkSecurityGroupIds
			sgLocal := typeSetToStringArray(d.Get("security_groups").(*schema.Set))

			if len(sgLocal) == RESOURCE_EMPTY && len(sgRemote) > 1 {
				if errSetSg := d.Set("security_groups", resp.Result.NetworkInterface.NetworkSecurityGroupIds[1:]); errSetSg != nil {
					return resource.NonRetryableError(formatArraySetErrorMessage(errSetSg))
				}
			}

			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND {
			log.Printf("Resource not found, probably have been deleted")
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

func resourceJDCloudNetworkInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if d.HasChange("network_interface_name") || d.HasChange("security_groups") || d.HasChange("description") {

		config := meta.(*JDCloudConfig)
		vpcClient := client.NewVpcClient(config.Credential)

		req := apis.NewModifyNetworkInterfaceRequestWithAllParams(
			config.Region,
			d.Id(),
			GetStringAddr(d, "network_interface_name"),
			GetStringAddr(d, "description"),
			typeSetToStringArray(d.Get("security_groups").(*schema.Set)))
		resp, err := vpcClient.ModifyNetworkInterface(req)

		if err != nil {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceUpdate failed %s ", err.Error())
		}

		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceUpdate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
		}

		d.SetPartial("network_interface_name")
		d.SetPartial("security_groups")
		d.SetPartial("description")
	}

	if d.HasChange("secondary_ip_addresses") || d.HasChange("secondary_ip_count") {

		pInterface, cInterface := d.GetChange("secondary_ip_addresses")
		p := pInterface.(*schema.Set)
		c := cInterface.(*schema.Set)
		i := p.Intersection(c)

		if err := performSecondaryIpDetach(d, meta, p.Difference(i)); len(typeSetToStringArray(p.Difference(i))) != 0 && err != nil {
			return err
		}
		if err := performSecondaryIpAttach(d, meta, c.Difference(i), d.Get("secondary_ip_count").(int)); len(typeSetToStringArray(c.Difference(i))) != 0 && err != nil {
			return err
		}

		d.SetPartial("secondary_ip_addresses")
		d.SetPartial("secondary_ip_count")
	}

	d.Partial(false)
	return nil
}

func resourceJDCloudNetworkInterfaceDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	rq := apis.NewDeleteNetworkInterfaceRequest(config.Region, d.Get("network_interface_id").(string))
	resp, err := vpcClient.DeleteNetworkInterface(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkInterfaceDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.SetId("")
	return nil
}

func performSecondaryIpDetach(d *schema.ResourceData, m interface{}, set *schema.Set) error {

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewUnassignSecondaryIpsRequestWithAllParams(config.Region, d.Id(), typeSetToStringArray(set))
	resp, err := vpcClient.UnassignSecondaryIps(req)

	if err != nil {
		return fmt.Errorf("[ERROR] performSecondaryIpDetach Failed,reasons:%s", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] performSecondaryIpDetach failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	return nil
}

func performSecondaryIpAttach(d *schema.ResourceData, m interface{}, set *schema.Set, count int) error {

	force_tag := true

	config := m.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	req := apis.NewAssignSecondaryIpsRequestWithAllParams(config.Region, d.Id(), &force_tag, typeSetToStringArray(set), &count)
	resp, err := vpcClient.AssignSecondaryIps(req)

	if err != nil {
		return fmt.Errorf("[ERROR] performSecondaryIpAttach Failed,reasons:%s", err.Error())
	}
	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] performSecondaryIpAttach failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	return nil
}

func ipList(v []models.NetworkInterfacePrivateIp) []string {

	ipList := []string{}

	for _, vv := range v {

		ipList = append(ipList, vv.PrivateIpAddress)
	}
	return ipList
}
