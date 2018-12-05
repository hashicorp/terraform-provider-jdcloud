Terraform Provider for JDCloud
==================

[![Build Status](https://travis-ci.com/jdclouddevelopers/terraform-provider-jdcloud.svg?branch=master)](https://travis-ci.com/jdclouddevelopers/terraform-provider-jdcloud)


# Installation Guide

Terraform will **create/read/update/delete** resource on JDCloud for you automatically. Following guides
show you how to install Terraform together with JDCloud plugin. Commands are given under **Ubuntu**

## Build from binary

Download the Terraform binary, make sure Terraform binary is available in your `PATH`.
Download JDCloud plugin into the same directory as Terraform. Detailed instruction can be found in [release page](https://github.com/XiaohanLiang/terraform-provider-jdcloud/releases/edit/v0.1-beta)

## Build from source code (Recommended for developers)

### Clone this repository

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers
$ cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:jdclouddevelopers/terraform-provider-jdcloud
```

### Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/jdclouddevelopers/terraform-provider-jdcloud
$ make build
```

___
# Tutorial

## Provide access key and secret key

JDCloud resources was managed through a configuration namely "jdcloud.tf" placed in the same directory as Terraform
This tutorial gives you an **idea on how to edit this configuration file and how this plugin works**.
Before everything starts you have to provide a pair of [access key](https://docs.jdcloud.com/cn/account-management/accesskey-management).

```bash
provider "jdcloud" {
  access_key = "E1AD46FF7994BC3DF"
  secret_key = "B527396D788ABCDEF"
  region = "cn-north-1"
}
```

## Create a VPC resource through Terraform
VPC resource can be created by specifying the name of this VPC resource and the CIDR block. Meanwhile description on this resource is optional. Edit jdcloud.tf and then execute terraform apply. Resource on the cloud will be modified
```bash
resource "jdcloud_vpc" "vpc-1" {
  vpc_name = "my-vpc-1"
  cidr_block = "172.16.0.0/20"
  description = "example"
}
```
## Modify resource attributes through Terraform
 Just like creating them through console on web page. You can modify some attributes of resource after it has been created. Execute terraform applyafter it has been modified
```bash
resource "jdcloud_vpc" "vpc-1" {
  vpc_name = "my-vpc-1"
  cidr_block = "172.16.0.0/20"
  description = "new and modified description"
}
```
___
# More

**More example** on how to create a resource can be found [here](https://github.com/XiaohanLiang/terraform-provider-jdcloud/blob/master/example/main.tf).
**Restrictions** on creating/updating a resource can be found [here](https://docs.jdcloud.com/cn/).
**Terraform official** web page can be found [here](https://www.terraform.io/intro/index.html).
**Contact us JDCloud-Team** <ark@jd.com>

## License

Apache License Version 2


___
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