package jdcloud

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strings"
	"testing"
)

const TestAccOssFileConfig = `
resource "jdcloud_oss_bucket_upload" "devops" {
  bucket_name = "tffff"
  file_name = "/home/liangxiaohan/hello.cpp"
}`

func TestAccJDCloudOssFile_basic(t *testing.T) {

	var id, fileName string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccOssFileDestroy(&id, &fileName),
		Steps: []resource.TestStep{
			{
				Config: TestAccOssFileConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfOssFileExists("jdcloud_oss_bucket_upload.devops", &id, &fileName),
				),
			},
		},
	})
}

func testAccIfOssFileExists(name string, bucketName *string, fileName *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[name]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfOssFileExists Failed,we can not find a vpc namely:{%s} in terraform.State", name)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfOssFileExists Failed,operation failed, vpc is created but ID not set")
		}

		*bucketName = infoStoredLocally.Primary.Attributes["bucket_name"]
		fileNameFull := infoStoredLocally.Primary.Attributes["file_name"]
		fileNameParsed := strings.Split(fileNameFull, "/")
		*fileName = fileNameParsed[len(fileNameParsed)-1]

		svc := getOssClient(testAccProvider.Meta())
		resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(*bucketName)})
		if err != nil {
			return fmt.Errorf("[ERROR] Failed in testAccIfOssFileExists,reasons:%s", err.Error())
		}

		for _, item := range resp.Contents {
			if *fileName == *item.Key {
				return nil
			}
		}
		return fmt.Errorf("[ERROR] testAccIfOssFileExists Failed, File not found remotely")
	}
}

func testAccOssFileDestroy(bucketName, fileName *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *bucketName == "" || *fileName == "" {
			return fmt.Errorf("[ERROR] testAccOssFileDestroy Failed,id or fileName  is empty")
		}

		svc := getOssClient(testAccProvider.Meta())
		resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(*bucketName)})
		if err != nil {
			return fmt.Errorf("[ERROR] Failed in testAccOssFileDestroy,reasons:%s", err.Error())
		}

		for _, item := range resp.Contents {
			if *fileName == *item.Key {
				return fmt.Errorf("[ERROR] testAccOssFileDestroy Failed, File Not Deleted")
			}
		}
		return nil
	}
}
