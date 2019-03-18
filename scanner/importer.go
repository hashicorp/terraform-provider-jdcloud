package main

import (
	"fmt"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const filename = "jdcloud.tf"
const CONNECT_FAILED    = "Client.Timeout exceeded"
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
	if _, err := fd.Write(buf); err != nil {
		return
	}
	if err := fd.Close(); err != nil {
		return
	}
}

func connectionError(e error) bool {

	if e == nil {
		return false
	}
	ok, _ := regexp.MatchString(CONNECT_FAILED, e.Error())
	return ok
}

/*
	When request failed due to bad condition
	We use this function to generate a error
	With a formatted error Message
*/

func formatConnectionErrorMessage() error {

	pc, _, _, _ := runtime.Caller(1)
	nameFull := runtime.FuncForPC(pc).Name()
	nameEnd := filepath.Base(nameFull)
	funcName := strings.Split(nameEnd, ".")[1]

	template := ` [ERROR] Operation failed. Details are:

     + FunctionName :    %c[1;40;37m%s%c[0m 
     + Message:          Connection failed due to bad network condition`

	errorMessage := fmt.Sprintf(template, 0x1B, funcName, 0x1B)
	return fmt.Errorf(errorMessage)
}

/*
	When request failed due to other reasons ,
	For example : Resource name conflict, Incorrect parameters, etc.
	We use this function to generate a error with formatted error message
*/

func formatErrorMessage(respError core.ErrorResponse, e error) error {

	pc, _, line, _ := runtime.Caller(1)
	nameFull := runtime.FuncForPC(pc).Name()
	nameEnd := filepath.Base(nameFull)
	funcName := strings.Split(nameEnd, ".")[1]

	template := ` [ERROR] Operation failed. Details are:

     + FunctionName :    %c[1;40;37m%s%c[0m 
     + LineNumber :      %d
     + RequestError:     %#v 
     + Code:             %d 
     + Status:           %s
     + Message:          %s`

	errorMessage := fmt.Sprintf(template, 0x1B, funcName, 0x1B, line, e, respError.Code, respError.Status, respError.Message)
	return fmt.Errorf(errorMessage)
}


func main() {

	resourceMap = make(map[string]string)

	//copyVPC()
	//copySubnet()
	//copyRouteTable()
	//copySecurityGroup()
	//copyNetworkInterface()
	//copyEIP()
	//copyDisk()
	//copyInstance()
	copyRDS()
	//copyOSS()
}
