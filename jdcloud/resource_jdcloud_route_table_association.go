package jdcloud

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"time"
)

func resourceJDCloudRouteTableAssociation() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteTableAssociationCreate,
		Read:   resourceRouteTableAssociationRead,
		Update: resourceRouteTableAssociationUpdate,
		Delete: resourceRouteTableAssociationDelete,

		Schema: map[string]*schema.Schema{

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {

	attachList := typeSetToStringArray(d.Get("subnet_id").(*schema.Set))
	routeTableId := d.Get("route_table_id").(string)

	if err := performSubnetAttach(d, meta, attachList); err != nil {
		return err
	} else {
		d.SetId(routeTableId)
		return resourceRouteTableAssociationRead(d, meta)
	}
}

func resourceRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	associationClient := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := associationClient.DescribeRouteTable(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {

			if err := d.Set("subnet_id", resp.Result.RouteTable.SubnetIds); err != nil {
				return resource.NonRetryableError(formatArraySetErrorMessage(err))
			}
			return nil
		}

		if resp.Error.Code == RESOURCE_NOT_FOUND {
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

func resourceRouteTableAssociationUpdate(d *schema.ResourceData, m interface{}) error {

	d.Partial(true)
	defer d.Partial(false)

	if d.HasChange("subnet_id") {

		pInterface, cInterface := d.GetChange("subnet_id")
		p := pInterface.(*schema.Set)
		c := cInterface.(*schema.Set)
		i := p.Intersection(c)

		detachList := typeSetToStringArray(p.Difference(i))
		attachList := typeSetToStringArray(c.Difference(i))

		if err := performSubnetDetach(d, m, detachList); err != nil && len(detachList) != 0 {
			return err
		}
		d.SetPartial("subnet_id")

		if err := performSubnetAttach(d, m, attachList); err != nil && len(attachList) != 0 {
			return err
		}
		d.SetPartial("subnet_id")

	}

	return resourceRouteTableAssociationRead(d, m)
}

func resourceRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {

	subnetIds := typeSetToStringArray(d.Get("subnet_id").(*schema.Set))

	if err := performSubnetDetach(d, meta, subnetIds); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func typeSetToStringArray(set *schema.Set) []string {

	ret := []string{}
	for _, item := range set.List() {
		ret = append(ret, item.(string))
	}
	return ret
}

func performSubnetAttach(d *schema.ResourceData, meta interface{}, attachList []string) error {
	d.Partial(true)
	config := meta.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)
	req := apis.NewAssociateRouteTableRequest(config.Region, d.Get("route_table_id").(string), attachList)

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := disassociationClient.AssociateRouteTable(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetPartial("subnet_id")
			d.Partial(false)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func performSubnetDetach(d *schema.ResourceData, meta interface{}, detachList []string) error {

	config := meta.(*JDCloudConfig)
	disassociationClient := client.NewVpcClient(config.Credential)
	routeTableId := d.Get("route_table_id").(string)

	for _, id := range detachList {

		req := apis.NewDisassociateRouteTableRequest(config.Region, routeTableId, id)

		err := resource.Retry(time.Minute, func() *resource.RetryError {

			resp, err := disassociationClient.DisassociateRouteTable(req)
			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
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

	}

	return nil
}
