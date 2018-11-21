package jdcloud

import (
	"log"
	"os"
	"testing"
)

// TODO:Really wonder if this is necessary ?
// Since keys are used as API keys, however APIs are connected
// Since APIs are connected via m.(*JDConfig)
func testAccPreCheck(t *testing.T) {
	if accessKey := os.Getenv("JDCloud_ACCESS_KEY"); accessKey == "" {
		t.Fatalf("parameter : JDCloud_ACCESS_KEY must be set to complete testing")
	}
	if secretKey := os.Getenv("JDCloud_SECRET_KEY"); secretKey == "" {
		t.Fatalf("parameter : JDCloud_SECRET_KEY must be set to complete testing")
	}
	if regionID := os.Getenv("JDCloud_REGION"); regionID == "" {
		log.Println("region was not set, now they were set to cn-north-1")
		os.Setenv("JDCloud_REGION", "cn-north-1")
	}
}
