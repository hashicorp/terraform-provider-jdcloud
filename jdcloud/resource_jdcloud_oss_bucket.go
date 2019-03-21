package jdcloud

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"time"
)

const (
	jdcloudOssEndpoint = "s3.%s.jcloudcs.com"
)

func resourceJDCloudOssBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudOssBucketCreate,
		Read:   resourceJDCloudOssBucketRead,
		Update: resourceJDCloudOssBucketUpdate,
		Delete: resourceJDCloudOssBucketDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"acl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "private",
			},
		},
	}
}

func resourceJDCloudOssBucketCreate(d *schema.ResourceData, m interface{}) error {

	bucket := d.Get("bucket_name").(string)
	client := getOssClient(m)
	s3Input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	e := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := client.CreateBucket(s3Input)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "OperationAborted" {
				return resource.RetryableError(
					fmt.Errorf("[WARN] Error creating S3 bucket %s, retrying: %s", bucket, err))
			}
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if e != nil {
		return e
	}

	if _, err := client.PutBucketAcl(&s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(d.Get("acl").(string)),
	}); err != nil {
		return err
	}

	d.SetId(bucket)
	return resourceJDCloudOssBucketRead(d, m)
}

func resourceJDCloudOssBucketRead(d *schema.ResourceData, m interface{}) error {

	bucket := d.Id()
	client := getOssClient(m)
	s3Input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	if _, err := client.HeadBucket(s3Input); err != nil {
		if awsError, ok := err.(awserr.RequestFailure); ok && awsError.StatusCode() == RESOURCE_NOT_FOUND {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}
	return nil
}

func resourceJDCloudOssBucketUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("acl") {

		bucket := d.Get("bucket_name").(string)
		client := getOssClient(m)
		s3Input := &s3.PutBucketAclInput{
			Bucket: aws.String(bucket),
			ACL:    aws.String(d.Get("acl").(string)),
		}

		if _, err := client.PutBucketAcl(s3Input); err != nil {
			return err
		}

	}

	return resourceJDCloudOssBucketRead(d, m)
}

func resourceJDCloudOssBucketDelete(d *schema.ResourceData, m interface{}) error {
	bucket := d.Get("bucket_name").(string)
	client := getOssClient(m)
	s3Input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}

	if _, err := client.DeleteBucket(s3Input); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func getOssClient(m interface{}) *s3.S3 {
	config := m.(*JDCloudConfig)
	endpoint := fmt.Sprintf(jdcloudOssEndpoint, config.Region)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
			Region:      aws.String(config.Region),
			Endpoint:    aws.String(endpoint),
		},
	}))
	return s3.New(sess)
}
