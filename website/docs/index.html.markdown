---
layout: "jdcloud"
page_title: "Introduction"
sidebar_current: "docs-jdcloud-index"
description: |-
        This is an introduction to JDCloud plugin and help users setting up their credentials.
---

# JDCloud

JDCloud provider helps managing resources on JDCloud. Before you start with this plugin, 
you have to provide a pair of access key and secret to identify yourself. 

-> Navigation bar on the left gives you a brief mind on how to manage resources on JDCloud 
that is currently available.

## Authentication

Credential consists of your key pairs and the region id, which is used for authentication. 
Currently you can set up your credential in two ways: 

- Simply write them in your configuration file
- Set them as environment variables

### Write them in your configuration file

For example, a credential can look like this. Place this at the beginning of `jdcloud.tf`

```hcl
provider "jdcloud" {
  access_key = "your_access_key"
  secret_key = "your_secret_key"
  region     = "cn-north-1"
}
```

### Set as environment variable

Or you can set them as environment variable via command line

```bash
$ export access_key="your_access_key"
$ export secret_key="your_secret_key"
$ export region="cn-north-1"
```
And leave the provider field blank in configuration file. Terraform will load them automatically.

```hcl
provider "jdcloud" {

}
```
