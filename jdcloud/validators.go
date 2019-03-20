package jdcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	vpcApis "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/apis"
	vpcClient "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"github.com/satori/go.uuid"
	"math/rand"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

func validateStringInSlice(validSlice []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, err []error) {
		v, ok := i.(string)
		if !ok {
			err = append(err, fmt.Errorf("type of %s must be string", k))
			return
		}

		for _, str := range validSlice {
			if v == str || (ignoreCase && strings.ToLower(v) == strings.ToLower(str)) {
				return
			}
		}

		err = append(err, fmt.Errorf("expected %s to be one of %v, got %s", k, validSlice, v))
		return
	}
}

func validateStringNoEmpty(v interface{}, k string) (s []string, errs []error) {

	value := v.(string)
	if len(value) < 1 {
		errs = append(errs, fmt.Errorf("%v can not be Empty characters.", k))
	}

	return
}

func diskClientTokenDefault() string {
	var clientToken string
	nonce, _ := uuid.NewV4()
	clientToken = nonce.String()
	clientToken = strings.Replace(clientToken, "-", "", -1)

	if len(clientToken) > 20 {
		clientToken = string([]byte(clientToken)[:20])
	}
	return clientToken
}

func verifyVPC(d *schema.ResourceData, m interface{}, vpc, subnet string) error {

	config := m.(*JDCloudConfig)
	subnetClient := vpcClient.NewVpcClient(config.Credential)

	req := vpcApis.NewDescribeSubnetRequest(config.Region, subnet)
	resp, err := subnetClient.DescribeSubnet(req)

	if err != nil {
		return fmt.Errorf("[ERROR] verifyVPC Failed, when creating RDS %s ", err.Error())
	}

	if resp.Error.Code != 0 {
		return fmt.Errorf("[ERROR] verifyVPC Failed, when creating RDS  code:%d staus:%s message:%s ", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	if resp.Result.Subnet.VpcId != vpc {
		return fmt.Errorf("[ERROR] verifyVPC Failed, vpc ID does not match")
	}
	return nil
}

/*
	This function parse the error message to Check
	If error is lead by bad network condition
*/

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

/*
	This function is used to format error message in setting
	in TypeList or *Schema.Set()
*/

func formatArraySetErrorMessage(e error) error {

	pc, _, _, _ := runtime.Caller(1)
	nameFull := runtime.FuncForPC(pc).Name()
	nameEnd := filepath.Base(nameFull)
	funcName := strings.Split(nameEnd, ".")[1]

	template := ` [ERROR] Operation failed in setting TypeList/Schema.Set. Details are:

     + FunctionName :    %c[1;40;37m%s%c[0m 
     + ErrorMessage :     %#v`

	errorMessage := fmt.Sprintf(template, 0x1B, funcName, 0x1B, e)
	return fmt.Errorf(errorMessage)
}

/*
	For some times, when attributes can not be modified,
	we would like to ignore these modification
*/

func ignoreModify(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func validateDiskSize() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {

		diskSize := v.(int)
		if diskSize < MIN_DISK_SIZE || diskSize > MAX_DISK_SIZE {
			errors = append(errors, fmt.Errorf("[ERROR] Valid disk size varies from 20~3000, yours: %#v", diskSize))
		}
		if diskSize%10 != 0 {
			errors = append(errors, fmt.Errorf("[ERROR] Valid disk size must be in multiples of [10], that is,10,20,30..."))
		}
		return
	}
}

func validateStringCandidates(c ...string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {

		target, ok := v.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("[ERROR] Your parameters seems invalid, Parameter has to be a string , Yours:%v", target))
			return
		}

		invalid := true
		for _, candidate := range c {
			if target == candidate {
				invalid = false
			}
		}
		if invalid {
			errors = append(errors, fmt.Errorf("[ERROR] Your parameters seems invalid, \n\n\t Candidates: %v,\n\t Yours:%v", c, target))
		}
		return
	}
}

func generateDiskIndex(i interface{}) int {
	return rand.Intn(100)
}

// Nothing special, we modified the refresh frequency here
// This function is used rather than resource.Retry since under some
// bad network condition 500 ms is too short for a request to return
func RetryWithParamsSpecified(delay, frequency, timeout time.Duration, f resource.RetryFunc) error {
	var resultErr error
	var resultErrMu sync.Mutex

	c := &resource.StateChangeConf{
		Delay:      delay,
		Pending:    []string{"retryableerror"},
		Target:     []string{"success"},
		Timeout:    timeout,
		MinTimeout: frequency,
		Refresh: func() (interface{}, string, error) {
			rerr := f()

			resultErrMu.Lock()
			defer resultErrMu.Unlock()

			if rerr == nil {
				resultErr = nil
				return 42, "success", nil
			}

			resultErr = rerr.Err

			if rerr.Retryable {
				return 42, "retryableerror", nil
			}
			return nil, "quit", rerr.Err
		},
	}

	_, waitErr := c.WaitForState()

	// Need to acquire the lock here to be able to avoid race using resultErr as
	// the return value
	resultErrMu.Lock()
	defer resultErrMu.Unlock()

	// resultErr may be nil because the wait timed out and resultErr was never
	// set; this is still an error
	if resultErr == nil {
		return waitErr
	}
	// resultErr takes precedence over waitErr if both are set because it is
	// more likely to be useful
	return resultErr
}

//func commonRetryFunc(t time.Duration,command func()(interface{},error)) error {
//	return resource.Retry(t,func() *resource.RetryError{
//
//		resp,err := command()
//		if err == nil &&
//	})
//}
