Terraform Provider for JDCloud
==================

[![Build Status](https://travis-ci.com/jdclouddevelopers/terraform-provider-jdcloud.svg?branch=master)](https://travis-ci.com/jdclouddevelopers/terraform-provider-jdcloud)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/jdclouddevelopers/terraform-provider-jdcloud`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers
$ cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:jdclouddevelopers/terraform-provider-jdcloud
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/jdclouddevelopers/terraform-provider-jdcloud
$ make build
```

Contact
---------------------

[JDCloud-Team](ark@jd.com)


License
---------------------

Apache License Version 2


## Finish Resource:

- [ ] Instance
- [x] keypairs
- [ ] Disk
- [ ] EIP