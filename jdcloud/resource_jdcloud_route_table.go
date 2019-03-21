package jdcloud

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"time"
)

func resourceJDCloudRouteTable() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteTableCreate,
		Read:   resourceRouteTableRead,
		Update: resourceRouteTableUpdate,
		Delete: resourceRouteTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{

			"route_table_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceRouteTableCreate(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	conn := client.NewVpcClient(config.Credential)

	req := apis.NewCreateRouteTableRequestWithAllParams(config.Region,
		d.Get("vpc_id").(string),
		d.Get("route_table_name").(string),
		GetStringAddr(d, "description"))

	e := resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := conn.CreateRouteTable(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.RouteTableId)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})

	if e != nil {
		return e
	}
	return resourceRouteTableRead(d, m)
}

func resourceRouteTableRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	conn := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeRouteTableRequest(config.Region, d.Id())

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := conn.DescribeRouteTable(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.Set("route_table_name", resp.Result.RouteTable.RouteTableName)
			d.Set("description", resp.Result.RouteTable.Description)
			d.Set("vpc_id", resp.Result.RouteTable.VpcId)
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

func resourceRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	config := meta.(*JDCloudConfig)
	conn := client.NewVpcClient(config.Credential)

	if d.HasChange("route_table_name") || d.HasChange("description") {

		req := apis.NewModifyRouteTableRequestWithAllParams(config.Region,
			d.Id(),
			GetStringAddr(d, "route_table_name"),
			GetStringAddr(d, "description"))

		err := resource.Retry(time.Minute, func() *resource.RetryError {

			resp, err := conn.ModifyRouteTable(req)

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

		d.SetPartial("route_table_name")
		d.SetPartial("description")
	}
	d.Partial(false)
	return resourceRouteTableRead(d, meta)
}

func resourceRouteTableDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	conn := client.NewVpcClient(config.Credential)

	req := apis.NewDeleteRouteTableRequest(config.Region, d.Id())

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := conn.DeleteRouteTable(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId("")
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
