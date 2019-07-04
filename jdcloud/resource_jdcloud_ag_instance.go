package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	common "github.com/jdcloud-api/jdcloud-sdk-go/services/common/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/models"
	"log"
	"time"
)

/*
	Test Case:
		[]single ag, single template id
			delete it
		[]single ag, dual template id
			delete it
*/

func resourceJDCloudAGInstance() *schema.Resource {

	return &schema.Resource{
		Create: resourceJDCloudAGInstanceCreate,
		Read:   resourceJDCloudAGInstanceRead,
		Update: resourceJDCloudAGInstanceUpdate,
		Delete: resourceJDCloudAGInstanceDelete,

		Schema: map[string]*schema.Schema{
			"availability_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instances": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "Created by Terraform",
						},
						"instance_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceJDCloudAGInstanceCreate(d *schema.ResourceData, m interface{}) error {

	agId := d.Get("availability_group_id").(string)
	if e := createAgInstances(d, m, agId, d.Get("instances").(*schema.Set)); e != nil {
		return e
	}
	d.SetId(agId)
	return resourceJDCloudAGInstanceRead(d, m)
}

func resourceJDCloudAGInstanceRead(d *schema.ResourceData, m interface{}) error {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)
	req := apis.NewDescribeInstancesRequest(config.Region)
	req.Filters = []common.Filter{
		common.Filter{
			Name:   "agId",
			Values: []string{d.Id()},
		},
	}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {

		resp, err := vmClient.DescribeInstances(req)
		if err == nil && resp.Error.Code == REQUEST_COMPLETED {

			specs := []map[string]interface{}{}
			for _, item := range resp.Result.Instances {
				specs = append(specs, map[string]interface{}{
					"instance_id":   item.InstanceId,
					"instance_name": item.InstanceName,
					"description":   item.Description,
				})
			}
			if e := d.Set("instances", specs); e != nil {
				return resource.NonRetryableError(e)
			}

			return nil
		}

		if connectionError(err) {
			return resource.RetryableError(err)
		} else {
			return resource.NonRetryableError(err)
		}
	})
}

func resourceJDCloudAGInstanceUpdate(d *schema.ResourceData, m interface{}) error {

	d.Partial(true)
	defer d.Partial(false)

	if d.HasChange("instances") {

		previousInterface, currentInterface := d.GetChange("instances")
		previousSet := previousInterface.(*schema.Set)
		currentSet := currentInterface.(*schema.Set)
		intersect := previousSet.Intersection(currentSet)

		// Perform deleting
		detach := previousSet.Difference(intersect)
		if len(detach.List()) > 0 {
			ids := getIdLists(detach)
			if e := deleteInstances(d, m, ids); e != nil {
				return fmt.Errorf("AGInstance Update Failed in detaching, %v", e)
			}
		}
		d.SetPartial("instances")

		// Perform attaching
		attach := currentSet.Difference(intersect)
		if len(attach.List()) > 0 {
			if e := createAgInstances(d, m, d.Id(), attach); e != nil {
				return e
			}
		}
		d.SetPartial("instances")
	}

	return resourceJDCloudAGInstanceRead(d, m)
}

// Level-0 Get Id lists
func getIdLists(set *schema.Set) (ids []string) {

	for _, item := range set.List() {
		maps := item.(map[string]interface{})
		ids = append(ids, maps["instance_id"].(string))
	}
	return
}

// Level-0 Send requests only
func agInstancesSendRequests(m interface{}, reqs []*apis.CreateInstancesRequest) (instanceIds []string, errs []error) {

	config := m.(*JDCloudConfig)
	vmClient := client.NewVmClient(config.Credential)

	for _, req := range reqs {

		e := resource.Retry(2*time.Minute, func() *resource.RetryError {

			resp, err := vmClient.CreateInstances(req)
			if err == nil && resp.Error.Code == REQUEST_COMPLETED {
				instanceIds = append(instanceIds, resp.Result.InstanceIds[0])
				return nil
			}

			if connectionError(err) {
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(formatErrorMessage(resp.Error, err))
			}
		})
		if e != nil {
			errs = append(errs, e)
		}
	}
	return
}

// Level-0 Create some instances
func createAgInstances(d *schema.ResourceData, m interface{}, agId string, set *schema.Set) error {

	// Send some requests
	reqs := []*apis.CreateInstancesRequest{}
	for _, item := range set.List() {
		itemMap := item.(map[string]interface{})
		config := m.(*JDCloudConfig)
		req := apis.NewCreateInstancesRequest(config.Region, &vm.InstanceSpec{
			AgId:        &agId,
			Name:        itemMap["instance_name"].(string),
			Description: stringAddr(itemMap["description"].(string)),
		})
		reqs = append(reqs, req)
	}
	instanceIds, errs := agInstancesSendRequests(m, reqs)
	if len(errs) > 0 {
		return errs[0]
	}

	// Waiting until VMs are ready
	for _, instanceId := range instanceIds {
		if e := instanceStatusWaiter(d, m, instanceId, []string{VM_PENDING, VM_STARTING}, []string{VM_RUNNING}); e != nil {
			errs = append(errs, e)
		}

	}
	if len(errs) > 0 {
		log.Printf("Create Error happens returning...")
		return errs[0]
	}

	return nil
}

func resourceJDCloudAGInstanceDelete(d *schema.ResourceData, m interface{}) error {

	ids := getIdLists(d.Get("instances").(*schema.Set))

	if e := deleteInstances(d, m, ids); e != nil {
		return e
	}

	d.SetId("")
	return nil
}
