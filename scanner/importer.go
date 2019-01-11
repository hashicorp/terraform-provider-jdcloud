package main

import (
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"os"
)

const filename = "jdcloud.tf"

var (
	resourceMap map[string]string
	region      = os.Getenv("region")
	access_key  = os.Getenv("access_key")
	secret_key  = os.Getenv("secret_key")
	config      = &JDCloudConfig{
		AccessKey:  access_key,
		SecretKey:  secret_key,
		Region:     region,
		Credential: core.NewCredentials(access_key, secret_key),
	}
)

type JDCloudConfig struct {
	AccessKey  string
	SecretKey  string
	Region     string
	Credential *core.Credential
}

func tracefile(str_content string) {
	fd, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	buf := []byte(str_content)
	fd.Write(buf)
	fd.Close()
}

func main() {

	resourceMap = make(map[string]string)

	copyVPC()
	copySubnet()
	copyRouteTable()
	copySecurityGroup()
	copyNetworkInterface()
	copyEIP()
	copyDisk()
	copyInstance()
	copyRDS()
	copyOSS()
}
