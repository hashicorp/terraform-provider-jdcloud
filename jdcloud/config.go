package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
)

type (
	JDCloudConfig struct {
		AccessKey  string
		SecretKey  string
		Region     string
		Credential *core.Credential
	}
)

var (
	regionCn = map[string]string{
		"cn-north-1": "华北-北京",
		"cn-south-1": "华南-广州",
		"cn-east-1":  "华东-宿迁",
		"cn-east-2":  "华东-上海",
	}
)

func initConfig(d *schema.ResourceData) (interface{}, error) {
	region := d.Get("region").(string)
	if _, ok := regionCn[region]; !ok {
		return nil, fmt.Errorf("Invalid region '%s'", region)
	}

	conf := &JDCloudConfig{
		AccessKey: d.Get("access_key").(string),
		SecretKey: d.Get("secret_key").(string),
		Region:    region,
		Credential: core.NewCredentials(
			d.Get("access_key").(string),
			d.Get("secret_key").(string),
		),
	}
	return conf, nil
}
