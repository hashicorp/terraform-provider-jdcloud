package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	rds "github.com/jdcloud-api/jdcloud-sdk-go/services/rds/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/rds/client"
	"regexp"
	"time"
)

/*
	By modifying any attributes in database
	may lead to an rebuilding and data loss
*/

func resourceJDCloudRDSDatabase() *schema.Resource {

	return &schema.Resource{
		Create: resourceJDCloudRDSDatabaseCreate,
		Read:   resourceJDCloudRDSDatabaseRead,
		Delete: resourceJDCloudRDSDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"character_set": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJDCloudRDSDatabaseCreate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewCreateDatabaseRequest(config.Region, d.Get("instance_id").(string), d.Get("db_name").(string), d.Get("character_set").(string))

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.CreateDatabase(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.RequestID)
			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(formatConnectionErrorMessage())
		} else {
			return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
		}
	})
}

func resourceJDCloudRDSDatabaseRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)
	req := apis.NewDescribeDatabasesRequest(config.Region, d.Get("instance_id").(string))

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.DescribeDatabases(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.RequestID)
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

func resourceJDCloudRDSDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceJDCloudRDSDatabaseDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*JDCloudConfig)
	rdsClient := client.NewRdsClient(config.Credential)

	req := apis.NewDeleteDatabaseRequest(config.Region, d.Get("instance_id").(string), d.Get("db_name").(string))

	return resource.Retry(time.Minute, func() *resource.RetryError {

		resp, err := rdsClient.DeleteDatabase(req)

		if err == nil && resp.Error.Code == REQUEST_COMPLETED {
			d.SetId(resp.RequestID)
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
