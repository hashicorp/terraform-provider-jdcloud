package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	rds "github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"regexp"
	"time"
)

func resourceJDCloudRDSDatabase() *schema.Resource {

	return &schema.Resource{
		Create: resourceJDCloudRDSDatabaseCreate,
		Read:   resourceJDCloudRDSDatabaseRead,
		Update: resourceJDCloudRDSDatabaseUpdate,
		Delete: resourceJDCloudRDSDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"db_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"character_set": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceJDCloudRDSDatabaseCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewCreateDatabaseRequest(config.Region, d.Get("instance_id").(string), d.Get("db_name").(string), d.Get("character_set").(string))
	resp, err := rdsClient.CreateDatabase(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDatabaseCreate failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDatabaseCreate failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	d.SetId(resp.RequestID)
	return nil
}

func resourceJDCloudRDSDatabaseRead(d *schema.ResourceData, meta interface{}) error {

	resp, err := keepReading(d.Get("instance_id").(string), meta)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDatabaseRead failed %s ", err.Error())
	}

	if resp.Error.Code == RESOURCE_NOT_FOUND {
		d.SetId("")
		return nil
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDatabaseRead failed  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	for _, db := range resp.Result.Databases {
		if d.Get("db_name").(string) == db.DbName {
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceJDCloudRDSDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("instance_id") || d.HasChange("db_name") || d.HasChange("character_set") {
		originId, _ := d.GetChange("instance_id")
		originDb, _ := d.GetChange("db_name")
		originSet, _ := d.GetChange("character_set")
		d.Set("instance_id", originId)
		d.Set("db_name", originDb)
		d.Set("character_set", originSet)
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDatabaseUpdate failed,Attributes cannot be modified")
	}

	return nil
}

func resourceJDCloudRDSDatabaseDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewDeleteDatabaseRequest(config.Region, d.Get("instance_id").(string), d.Get("db_name").(string))
	resp, err := rdsClient.DeleteDatabase(req)

	if err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDatabaseDelete failed %s ", err.Error())
	}

	if resp.Error.Code != REQUEST_COMPLETED {
		return fmt.Errorf("[ERROR] resourceJDCloudRDSDatabaseDelete failed  code:%d staus:%s message:%s ,result:%#v", resp.Error.Code, resp.Error.Status, resp.Error.Message, resp.Result)
	}

	return nil
}

// Reading Database often leads to connection error, this function will reconnect for at most 3 times
func keepReading(instanceId string, m interface{}) (*rds.DescribeDatabasesResponse, error) {

	config := m.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewDescribeDatabasesRequest(config.Region, instanceId)

	for count := 0; count < RDS_MAX_RECONNECT; count++ {

		resp, err := rdsClient.DescribeDatabases(req)

		if err == nil {
			return resp, err
		}

		if s, _ := regexp.MatchString(CONNECT_FAILED, err.Error()); s {
			time.Sleep(3 * time.Second)
			continue
		}

		return resp, err
	}

	return nil, fmt.Errorf("[ERROR] keepReading Failed, MAX_RECONNECT_EXCEDEEDED")
}
