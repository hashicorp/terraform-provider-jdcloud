package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	"testing"
)

const TestAccKeyPairsConfig = `
resource "jdcloud_key_pairs" "keypairs_1" {
  key_name   = "JDCLODU-123312FMK"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAm3c0aR7zI0Xkrm1MD3zDrazC+fR+DV6p/xdzQJWviqPSFGfsatptY3Bc6gYF/qY+Jjccmrm6SKrtW0xPicCw5/uGAuIyhzBnG1Ix0fITdJkeBzyBpxdu/oxnJuvu5P8BLfoFH80ovUqysnttC/7hHBp+uIctkt/g14Kqd7kuPc0Gp4cx7tntNWpmzHJI9i+ayF95nJyFGIjF/s57b9pcKnnv2LXkMDNxsnzgWwPpi2hqGpQSz//ji8GgSED08u34zSjVbPc0TYJy4GO+uD8hozGnf9Evlpqx4OSB0D+4AuRcIniPgCOotYpOdp3Lj7CQRwzkiFZ6YpOxj1qMD4fnjQ== rsa-key-jddevelop"
}
`

func TestAccJDCloudKeyPairs_basic(t *testing.T) {

	var keyName string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccKeyPairsDestroy(&keyName),
		Steps: []resource.TestStep{
			{
				Config: TestAccKeyPairsConfig,
				Check: resource.ComposeTestCheckFunc(

					testAccIfKeyPairsExists("jdcloud_key_pairs.keypairs_1", &keyName),
					resource.TestCheckResourceAttr("jdcloud_key_pairs.keypairs_1", "key_name", "JDCLODU-123312FMK"),
					resource.TestCheckResourceAttrSet("jdcloud_key_pairs.keypairs_1", "key_finger_print"),
					resource.TestCheckResourceAttrSet("jdcloud_key_pairs.keypairs_1", "private_key"),
				),
			},
		},
	})
}

//-------------------------- Customized check functions

func testAccIfKeyPairsExists(name string, id *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		infoStoredLocally, ok := stateInfo.RootModule().Resources[name]
		if ok == false {
			return fmt.Errorf("[ERROR] testAccIfKeyPairsExists Failed,we can not find a resource namely:{%s} in terraform.State", name)
		}
		if infoStoredLocally.Primary.ID == "" {
			return fmt.Errorf("[ERROR] testAccIfKeyPairsExists Failed,operation failed, resource :{%s} is created but ID not set", name)
		}
		idStoredLocally := infoStoredLocally.Primary.Attributes["key_name"]

		config := testAccProvider.Meta().(*JDCloudConfig)
		clientKey := client.NewVmClient(config.Credential)

		req := apis.NewDescribeKeypairsRequest(config.Region)
		resp, err := clientKey.DescribeKeypairs(req)

		if err != nil {
			return err
		}
		if resp.Error.Code != REQUEST_COMPLETED {
			return fmt.Errorf("[ERROR] testAccIfKeyPairsExists Failed,invalid region id")
		}

		keysExists := false
		for _, key := range resp.Result.Keypairs {
			if idStoredLocally == key.KeyName {
				keysExists = true
				break
			}
		}
		if keysExists == false {
			return fmt.Errorf("[ERROR] testAccIfKeyPairsExists Failed,keys not been created remotely")
		}

		*id = idStoredLocally
		return nil
	}
}

func testAccKeyPairsDestroy(name *string) resource.TestCheckFunc {

	return func(stateInfo *terraform.State) error {

		if *name == "" {
			return fmt.Errorf("[ERROR] testAccKeyPairsDestroy Failed,name is empty")
		}

		config := testAccProvider.Meta().(*JDCloudConfig)
		clientKey := client.NewVmClient(config.Credential)

		req := apis.NewDescribeKeypairsRequest(config.Region)
		resp, err := clientKey.DescribeKeypairs(req)

		if err != nil {
			return err
		}
		keysExists := false
		for _, key := range resp.Result.Keypairs {
			if *name == key.KeyName {
				keysExists = true
				break
			}
		}
		if keysExists == true {
			return fmt.Errorf("[ERROR] testAccKeyPairsDestroy Failed,keys still exists")
		}
		return nil
	}
}
