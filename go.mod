module github.com/terraform-providers/terraform-provider-jdcloud

go 1.12

require (
	github.com/aws/aws-sdk-go v1.19.18
	github.com/hashicorp/terraform v0.12.0
	github.com/jdcloud-api/jdcloud-sdk-go v1.9.0
	github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid v1.2.0
)

replace github.com/jdclouddevelopers/terraform-provider-jdcloud => ./
