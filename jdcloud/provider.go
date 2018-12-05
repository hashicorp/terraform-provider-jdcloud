package jdcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"jdcloud_disk":                         resourceJDCloudDisk(),
			"jdcloud_oss_bucket":                   resourceJDCloudOssBucket(),
			"jdcloud_instance":                     resourceJDCloudInstance(),
			"jdcloud_key_pairs":                    resourceJDCloudKeyPairs(),
			"jdcloud_network_security_group":       resourceJDCloudNetworkSecurityGroup(),
			"jdcloud_network_security_group_rules": resourceJDCloudNetworkSecurityGroupRules(),
			"jdcloud_vpc":                          resourceJDCloudVpc(),
			"jdcloud_network_interface":            resourceJDCloudNetworkInterface(),
			"jdcloud_disk_attachment":              resourceJDCloudDiskAttachment(),
			"jdcloud_subnet":                       resourceJDCloudSubnet(),
			"jdcloud_route_table":                  resourceJDCloudRouteTable(),
			"jdcloud_route_table_association":      resourceJDCloudRouteTableAssociation(),
			"jdcloud_eip":                          resourceJDCloudEIP(),
			"jdcloud_eip_association":              resourceJDCloudAssociateElasticIp(),
			"jdcloud_route_table_rules":            resourceJDCloudRouteTableRules(),
			"jdcloud_network_interface_attachment": resourceJDCloudNetworkInterfaceAttach(),
			"jdcloud_route_table_rule":             resourceJDCloudRouteTableRule(),
			"jdcloud_rds_instance":                 resourceJDCloudRDSInstance(),
			"jdcloud_rds_account":                  resourceJDCloudRDSAccount(),
			"jdcloud_rds_database":                 resourceJDCloudRDSDatabase(),
			"jdcloud_rds_privilege":                resourceJDCloudRDSPrivilege(),
		},
		Schema: map[string]*schema.Schema{
			"access_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("access_key", nil),
				Description: "Access key for API operations",
			},
			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("secret_key", nil),
				Description: "Secret key for API operations",
			},
			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("region", nil),
				Description: "The region where JDCLOUD operations will take place",
			},
		},
		ConfigureFunc: initConfig,
	}
}
