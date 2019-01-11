package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const jdcloudOssEndpoint = "s3.%s.jcloudcs.com"
const ossTemplate = `
resource "jdcloud_oss_bucket" "%s" {
  bucket_name = "%s"
  acl = "%s"
}
`

var privMap = map[string]string{
	"READ":         "public-read",
	"FULL_CONTROL": "public-read-write",
}

func copyOSS() {

	s := getOssClient()
	o, e := s.ListBuckets(&s3.ListBucketsInput{})
	if e != nil {
		fmt.Printf("[ERROR] copyOSS()-ListBucket Failed,reasons: %s", e.Error())
	}

	for index, bucket := range o.Buckets {

		fmt.Println(*bucket.Name)
		result, err := s.GetBucketAcl(&s3.GetBucketAclInput{Bucket: bucket.Name})

		resourceName := fmt.Sprintf("oss-%d", index)
		if err != nil {
			tracefile(fmt.Sprintf(ossTemplate, resourceName, *bucket.Name, "private"))
		} else {
			tracefile(fmt.Sprintf(ossTemplate, resourceName, *bucket.Name, privMap[*result.Grants[0].Permission]))
		}
	}
}

func getOssClient() *s3.S3 {
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
