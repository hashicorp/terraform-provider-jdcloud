---
layout: "jdcloud"
page_title: "JDCloud Key Pairs"
sidebar_current: "docs-jdcloud-resource-key-pairs"
description: |-
  Provides a JDCloud Key Pairs
---

# jdcloud\_key\_pairs

key\_pairs can be useful when trying to login to instance. This function helps to create key\_pairs

### Example Usage 

```hcl
resource "jdcloud_key_pairs" "keypairs_1" {
  key_name   = "JDCLOUD-KEY-PAIR"
  public_key = "ssh-rsa <Your Public Key> rsa-key-jddevelop"
}
```

### Argument Reference

The following arguments are supported:

* `key_name` - \(Required\): Name of your key pairs
* `public_key` - \(Optional\): There are two ways to upload keys. The trivial one is to upload your public\_key.
* `key_file` - \(Optional\): The other way is to provide your key\_file and the `private_key` will be returned

### Attributes Reference

The following attributes are exported:

* `key_finger_print` :  The finger print of this key pairs
* `private_key` :  When you try to generate keys, private key will be returned from server as response, stored locally.


