## 1.1.0 (Unreleased)
## 0.0.1 (March 27, 2019)

FEATURES:

* Now `InstanceTemplate` can be created with ssh keys. Tests,website added [GH-3]
* `Availability-Group` and `Instance-Template` ([#28](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/28))
* Updated to Go1.12, `go.mod` and `go.sum` has been generated [[#28](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/28)] 
* Scanner, back up your current infrastructure and save into a configuration file ([#21](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/21))
* Importer for various resources is introduced ([#17](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/17))
* Website files prepared ([#18](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/18))
* CHANGELOG.md created ([#18](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/18))
* Compatible to Terraform0.12([#7](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/7))

BUG FIXES:

* Multiple bugs fixed in `Scanner` ([#28](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/28))
* Field `minimum_amount` is set for *Schema.TypeSet ([#22](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/22))
* When updating *Schema.TypeSet field and failed on the halfway, `SetPartial` is added to correctly modify `tfstate`([#22](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/22))
* Unable to set ACL when OSS bucket is created ([#17](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/17))
* No validation on `subnet_id` and `vpc_id` when creating RDS instance([#15](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/15))

IMPROVEMENTS:

* Retry function is introduced to avoid bad network condition issues([#22](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/22))
* `Oss` and `OssUpload` test file has been impleted([#22](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/22))
* Code format modified. Errors will be returned when trying to invoke `d.Set("List/Set",List/Set)` ([#17](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/17))
* `*schema.TypeList` -> `*schema.TypeSet` ([#17](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/17))
* GNUMakefile introduced ([#17](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/17))
* Scripts for various purpose introduced [[#18](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/18)] 
* .travis.yml modified for a more detailed testing process ([#18](https://github.com/terraform-providers/terraform-provider-jdcloud/issues/18))
