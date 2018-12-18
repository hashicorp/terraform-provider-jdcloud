package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"log"
)

func resourceJDCloudNetworkAcl() *schema.Resource {

	return &schema.Resource{

		Create: resourceJDCloudNetworkAclCreate,
		Read:   resourceJDCloudNetworkAclRead,
		Delete: resourceJDCloudNetworkAclDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceJDCloudNetworkAclCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	vpcId := d.Get("vpc_id").(string)
	networkAclName := d.Get("name").(string)

	req := apis.NewCreateNetworkAclRequest(config.Region, vpcId, networkAclName)

	if _, ok := d.GetOk("description"); ok {
		req.Description = GetStringAddr(d, "description")
	}

	resp, err := vpcClient.CreateNetworkAcl(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkAclCreate failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkAclCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}
	d.SetId(resp.Result.NetworkAclId)

	return nil
}

func resourceJDCloudNetworkAclRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)
	req := apis.NewDescribeNetworkAclRequest(config.Region, d.Id())
	resp, err := vpcClient.DescribeNetworkAcl(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkAclRead failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		log.Printf("Resource not found, probably have been deleted")
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] failed in resourceJDCloudNetworkAclRead code:%d message:%s", resp.Error.Code, resp.Error.Message)
	}

	d.Set("name", resp.Result.NetworkAcl.NetworkAclName)
	d.Set("vpc_id", resp.Result.NetworkAcl.VpcId)
	d.Set("description", resp.Result.NetworkAcl.Description)
	d.Set("created_time", resp.Result.NetworkAcl.CreatedTime)

	return nil
}

func resourceJDCloudNetworkAclDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*JDCloudConfig)
	vpcClient := client.NewVpcClient(config.Credential)

	networkAclId := d.Id()
	rq := apis.NewDeleteNetworkAclRequest(config.Region, networkAclId)
	resp, err := vpcClient.DeleteNetworkAcl(rq)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkAclDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudNetworkAclDelete failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}
