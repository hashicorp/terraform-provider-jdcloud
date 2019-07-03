package jdcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"testing"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]terraform.ResourceProvider
var packer_subnet = "subnet-rht03mi6o0"
var	packer_vpc = "vpc-z9q9xwmb1d"
var	packer_image = "img-m82soo3mes"
var	packer_sg = "sg-s0ardxmz3a"
var	packer_sgs= `["sg-s0ardxmz3a","sg-js9if78wqp"]`

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]terraform.ResourceProvider{
		"jdcloud": testAccProvider,
	}
}

// This step is necessary since we need to pass the
// Secret key and public key to begin our testing
func testAccPreCheck(t *testing.T) {
	if accessKey := os.Getenv("access_key"); accessKey == "" {
		t.Fatalf("parameter : access_key must be set to complete testing")
	}
	if secretKey := os.Getenv("secret_key"); secretKey == "" {
		t.Fatalf("parameter : secret_key must be set to complete testing")
	}
	if regionID := os.Getenv("region"); regionID == "" {
		log.Println("region was not set, now they were set to cn-north-1")
		os.Setenv("region", "cn-north-1")
	}
}
