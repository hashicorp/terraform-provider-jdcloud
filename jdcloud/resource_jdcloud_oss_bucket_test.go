package jdcloud

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

const TestAccOssConfig = `
resource "jdcloud_oss_bucket" "jd-bucket-2" {
  bucket_name = "example"
  acl = "private"
}
`

func TestAccJDCloudOss_basic(t *testing.T) {

	var id string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccOssDestroy(&id),
		Steps: []resource.TestStep{
			{
				Config: TestAccOssConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfOssExists("jdcloud_oss_bucket.jd-bucket-2", &id),
				),
			},
		},
	})
}

func testAccIfOssExists(name string, id *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[name]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfOssExists Failed,we can not find a oss namely:{%s} in terraform.State", name)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfOssExists Failed,operation failed, oss is created but ID not set")
		}
		*id = infoStoredLocally.Primary.ID

		conn := getOssClient(testAccProvider.Meta())
		s3Input := &s3.HeadBucketInput{
			Bucket: aws.String(*id),
		}

		if _, err := conn.HeadBucket(s3Input); err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchBucket {
				return nil
			} else {
				return err
			}
		}
		return nil
	}
}

func testAccOssDestroy(id *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *id == "" {
			return fmt.Errorf("[ERROR] testAccOssDestroy Failed,OssID is empty")
		}

		conn := getOssClient(testAccProvider.Meta())
		s3Input := &s3.HeadBucketInput{
			Bucket: aws.String(*id),
		}

		if _, err := conn.HeadBucket(s3Input); err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchBucket {
				return err
			} else {
				return nil
			}
		}
		return fmt.Errorf("[ERROR] testAccOssDestroy failed, position-1")
	}
}
