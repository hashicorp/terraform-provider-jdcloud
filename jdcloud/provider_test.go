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

const packer_subnet = "subnet-rht03mi6o0"
const packer_vpc = "vpc-z9q9xwmb1d"
const packer_image = "img-m82soo3mes"
const packer_sg = "sg-s0ardxmz3a"
const packer_sgs = `["sg-s0ardxmz3a","sg-js9if78wqp"]`
const packer_template = "it-kn5ok4o4hi"
const packer_ag = "ag-6mtp6pa11v"
const packer_instance = "i-8yi4jyr273"
const packer_disk = "vol-9dya7e5rdi"
const packer_disk2 = "vol-qm7t7q7pmk"
const packer_eip = "fip-jr6szry26y"
const packer_eip2 = "fip-bjvpenb8ux"
const packer_interface = "port-2phhnk57rw"
const packer_rds = "mysql-155pjskhpy"
const packer_route = "rtb-jgso5x1ein"
const packer_route_association = "rtb-elbnrrw3tu"

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
