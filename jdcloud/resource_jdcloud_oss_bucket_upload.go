package jdcloud

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/terraform/helper/schema"
	"os"
	"path/filepath"
	"strings"
)

func resourceJDCloudOssBucketUpload() *schema.Resource {
	return &schema.Resource{
		Create: resourceJDCloudOssBucketUploadCreate,
		Read:   resourceJDCloudOssBucketUploadRead,
		Delete: resourceJDCloudOssBucketUploadDelete,

		Schema: map[string]*schema.Schema{
			"bucket_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"file_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"remote_location": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func getUploader(meta interface{}) *s3manager.Uploader {
	config := meta.(*JDCloudConfig)
	endpoint := fmt.Sprintf(jdcloudOssEndpoint, config.Region)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
			Region:      aws.String(config.Region),
			Endpoint:    aws.String(endpoint),
		},
	}))
	return s3manager.NewUploader(sess)
}

func resourceJDCloudOssBucketUploadCreate(d *schema.ResourceData, meta interface{}) error {

	bucketName := d.Get("bucket_name").(string)
	fileName := d.Get("file_name").(string)

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to open file namely %s, Error message-%s", fileName, err)
	}
	defer file.Close()

	uploader := getUploader(meta)
	respUpload, errUpload := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath.Base(fileName)),
		Body:   file,
	})
	if errUpload != nil || respUpload.Location == "" {
		return fmt.Errorf("[ERROR] Failed to upload file: %s", errUpload.Error())
	}

	d.Set("remote_location", respUpload.Location)
	d.SetId(respUpload.Location)
	return nil
}

func resourceJDCloudOssBucketUploadRead(d *schema.ResourceData, meta interface{}) error {

	svc := getOssClient(meta)
	bucketName := d.Get("bucket_name").(string)
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucketName)})
	if err != nil {
		return fmt.Errorf("[ERROR] Failed in resourceJDCloudOssBucketUploadRead,reasons:%s", err.Error())
	}

	fileNameFull := d.Get("file_name").(string)
	fileNameParsed := strings.Split(fileNameFull, "/")
	fileName := fileNameParsed[len(fileNameParsed)-1]

	for _, item := range resp.Contents {
		if fileName == *item.Key {
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceJDCloudOssBucketUploadDelete(d *schema.ResourceData, meta interface{}) error {

	bucketName := d.Get("bucket_name").(string)
	remoteLocation := d.Get("remote_location").(string)

	client := getOssClient(meta)
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath.Base(remoteLocation)),
	})

	if err != nil {
		return fmt.Errorf("[ERROR] Failed in deleting OSS-Object,Error message:%s", err)
	}
	d.SetId("")
	return nil
}
