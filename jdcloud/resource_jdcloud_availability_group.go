package jdcloud

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/ag/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/ag/client"
	"log"
	"time"
)

func resourceJDCloudAvailabilityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudAvailabilityGroupCreate,
		Read:   resourceJDCloudAvailabilityGroupRead,
		Update: resourceJDCloudAvailabilityGroupUpdate,
		Delete: resourceJDCloudAvailabilityGroupDelete,

		Schema: map[string]*schema.Schema{
			"availability_group_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"az": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
				ForceNew: true,
			},
			"instance_template_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ag_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "kvm",
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceJDCloudAvailabilityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*JDCloudConfig)

	req := apis.NewCreateAgRequest(config.Region)
	req.SetAzs(typeSetToStringArray(d.Get("az").(*schema.Set)))
	req.SetAgName(d.Get("availability_group_name").(string))
	req.SetInstanceTemplateId(d.Get("instance_template_id").(string))
	req.SetAgType(d.Get("ag_type").(string))
	if _, ok := d.GetOk("description"); ok {
		req.SetDescription(d.Get("description").(string))
	}

	agClient := client.NewAgClient(config.Credential)
	log.Printf("nishizhu 1")
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := agClient.CreateAg(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.Result.AgId)
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
	log.Printf("nishizhu 1.1")
	return resourceJDCloudAvailabilityGroupRead(d, meta)
}

func resourceJDCloudAvailabilityGroupDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	req := apis.NewDeleteAgRequest(config.Region, d.Id())
	agClient := client.NewAgClient(config.Credential)
	log.Printf("nishizhu 1")
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := agClient.DeleteAg(req)
		log.Printf("nishizhu 2 %v", resp)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId("")
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

	return resourceJDCloudAvailabilityGroupRead(d, meta)
}

func resourceJDCloudAvailabilityGroupRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("nishizhu 2")
	config := meta.(*JDCloudConfig)
	req := apis.NewDescribeAgRequest(config.Region, d.Id())
	agClient := client.NewAgClient(config.Credential)

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := agClient.DescribeAg(req)
		log.Printf("nishizhu 2.1 %v", resp.Result.Ag)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {

			d.Set("instance_template_id", resp.Result.Ag.InstanceTemplateId)
			d.Set("ag_type", resp.Result.Ag.AgType)
			d.Set("availability_group_name", resp.Result.Ag.Name)
			d.Set("description", resp.Result.Ag.Description)
			if e := d.Set("az", resp.Result.Ag.Azs); e != nil {
				return resource.NonRetryableError(e)
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

	if err != nil {
		return err
	}

	return nil
}

func resourceJDCloudAvailabilityGroupUpdate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)

	if d.HasChange("availability_group_name") || d.HasChange("description") {

		req := apis.NewUpdateAgRequestWithAllParams(config.Region, d.Id(), stringAddr(d.Get("description")), stringAddr(d.Get("availability_group_name")))
		agClient := client.NewAgClient(config.Credential)

		err := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := agClient.UpdateAg(req)

			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
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

		if err != nil {
			return err
		}
	}

	return resourceJDCloudAvailabilityGroupRead(d, meta)
}
