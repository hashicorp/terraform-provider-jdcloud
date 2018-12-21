Terraform Provider for JDCloud
==================

[![Build Status](https://travis-ci.com/jdclouddevelopers/terraform-provider-jdcloud.svg?branch=master)](https://travis-ci.com/jdclouddevelopers/terraform-provider-jdcloud)


# Installation Guide

Terraform will **create/read/update/delete** resource on JDCloud for you automatically. Following guides
show you how to install Terraform together with JDCloud plugin. Commands are given under **Ubuntu**
<br>

### Build from binary

Download the Terraform binary, make sure Terraform binary is available in your `PATH`.
Download JDCloud plugin into the same directory as Terraform. Detailed instruction can be found in [release page](https://github.com/jdclouddevelopers/terraform-provider-jdcloud/releases/edit/v0.1-beta)
<br>

### Build from source code (Recommended for developers)

**Clone this repository**

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers
$ cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:jdclouddevelopers/terraform-provider-jdcloud
```

**Enter the provider directory and build the provider**

```sh
$ cd $GOPATH/src/github.com/jdclouddevelopers/terraform-provider-jdcloud
$ make build
```
<br>

# Tutorial

### Provide access key and secret key

JDCloud resources were managed through a configuration file namely "jdcloud.tf", placed in the same directory as Terraform.
This tutorial gives you an **idea on how to edit this configuration file and how this plugin works**.
Before everything starts you have to provide [access key and secret key](https://docs.jdcloud.com/cn/account-management/accesskey-management).

```hcl
provider "jdcloud" {
  access_key = "EXAMPLEACCESSKEY"
  secret_key = "EXAMPLESECRETKEY"
  region = "cn-north-1"
}
```
<br>

### Create a VPC resource through Terraform
VPC resource can be created by specifying the name of this VPC resource and the CIDR block. Meanwhile description on this resource is optional. Edit `jdcloud.tf` and then execute `terraform apply`. Resource on the cloud will be modified consequently.
```hcl
resource "jdcloud_vpc" "vpc-1" {
  vpc_name = "my-vpc-1"
  cidr_block = "172.16.0.0/20"
  description = "example"
}
```
<br>

### Modify resource attributes through Terraform
Just like creating them through console on web page. You can modify some attributes of resource after it has been created. Execute terraform applyafter it has been modified
```hcl
resource "jdcloud_vpc" "vpc-1" {
  vpc_name = "my-vpc-1"
  cidr_block = "172.16.0.0/20"
  description = "new and modified description"
}
```
<br>



# More
**More example** on how to create a resource can be found [here](https://github.com/jdclouddevelopers/terraform-provider-jdcloud/blob/master/example/main.tf).  
**Restrictions** on creating/updating a resource can be found [here](https://docs.jdcloud.com/).  
**Terraform official** web page can be found [here](https://www.terraform.io/intro/index.html).  
**Contact us JDCloud-Team** <devops@jd.com>

## License

Mozilla Public License 2.0


<br>

# Finished Resource:

Terraform-JDCloud plugin is currently under developing, available resources are listed
below. Leave your feedback and advices for this plugin. Advices are very welcomed.

- [x] Instance
- [x] keyPairs
- [x] Disk
- [x] EIP
- [x] Network Interface
- [x] Security Group
- [x] RDS Cloud Database
- [x] Route Table
- [x] Subnet
- [x] VPC
