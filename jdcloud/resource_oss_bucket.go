package jdcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceOssBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceOssBucketCreate,
		Read:   resourceOssBucketRead,
		Update: resourceOssBucketUpdate,
		Delete: resourceOssBucketDelete,

		Schema: map[string]*schema.Schema{
			"bucket": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"acl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "private",
			},
		},
	}
}

func resourceOssBucketCreate(d *schema.ResourceData, m interface{}) error {
	bucket := d.Get("bucket").(string)
	d.SetId(bucket)
	return nil
}

func resourceOssBucketRead(d *schema.ResourceData, m interface{}) error {
	d.Id()

	if false {
		d.SetId("")
		return nil
	}

	d.Set("bucket", "xxx")
	return nil
}

func resourceOssBucketUpdate(d *schema.ResourceData, m interface{}) error {
	// Enable partial state mode
	d.Partial(true)

	if d.HasChange("acl") {
		// Try updating the address
		if err := updateAddress(d, m); err != nil {
			return err
		}

		d.SetPartial("acl")
	}

	// If we were to return here, before disabling partial mode below,
	// then only the "address" field would be saved.

	// We succeeded, disable partial mode. This causes Terraform to save
	// all fields again.
	d.Partial(false)

	return nil
}

func resourceOssBucketDelete(d *schema.ResourceData, m interface{}) error {
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return nil
}
