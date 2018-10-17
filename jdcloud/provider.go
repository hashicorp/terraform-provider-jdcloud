package jdcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"jdcloud_oss_bucket": resourceOssBucket(),
			"jdcloud_instance":   resourceJdcloudInstance(),
		},
		Schema: map[string]*schema.Schema{
			"access_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Access key for API operations",
			},
			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Secret key for API operations",
			},
			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The region where JDCLOUD operations will take place",
			},
		},
		ConfigureFunc: initConfig,
	}
}
