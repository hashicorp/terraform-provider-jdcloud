module terraform-provider-jdcloud

go 1.12

require (
	github.com/aws/aws-sdk-go v1.18.2
	github.com/hashicorp/terraform v0.11.13
	github.com/jdcloud-api/jdcloud-sdk-go v1.5.0
	github.com/jdclouddevelopers/terraform-provider-jdcloud v0.0.0-20190312095238-e8421309c082
	github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid v1.2.0
)

replace github.com/jdcloud-api/jdcloud-sdk-go => ./vendor/github.com/jdcloud-api/jdcloud-sdk-go/

replace github.com/satori/go.uuid => ./vendor/github.com/satori/go.uuid

replace github.com/jdclouddevelopers/terraform-provider-jdcloud => ./
