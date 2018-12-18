package jdcloud

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
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
			},
			"acl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "private",
			},
			"grant_full_control": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceJDCloudOssBucketCreate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)

	bucket := d.Get("bucket_name").(string)
	client := getOssClient(m)
	s3Input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	if _, err := client.CreateBucket(s3Input); err != nil {
		return fmt.Errorf("[ERROR] resourceJDCloudOssBucketCreate failed,Error message:%s", err.Error())
	}
	d.SetPartial("bucket_name")

	if _, err := client.PutBucketAcl(&s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(d.Get("acl").(string)),
	}); err != nil {
		return err
	}
	d.SetPartial("acl")

	d.SetId(bucket)

	d.Partial(true)
	return nil
}

func resourceJDCloudOssBucketRead(d *schema.ResourceData, m interface{}) error {

	bucket := d.Id()
	client := getOssClient(m)
	s3Input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	if _, err := client.HeadBucket(s3Input); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchBucket {
			d.SetId("")
		} else {
			return err
		}
	}

	return nil
}

func resourceJDCloudOssBucketUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)

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

		d.SetPartial("acl")
	}

	d.Partial(false)
	return nil
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
