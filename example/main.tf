provider "jdcloud" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region     = "${var.region}"
}

resource "jdcloud_oss_bucket" "jd-bucket-1" {
  bucket = "tf-test-b1"
  acl    = "public-read"
}

resource "jdcloud_oss_bucket" "jd-bucket-2" {
  bucket = "tf-test-b2"
}

resource "jdcloud_key_pairs" "key-1" {
  key_name = "key-1"
}

resource "jdcloud_vpc" "vpc-1" {
  vpc_name    = "my-vpc-1"
  cidr_block  = "172.16.0.0/20"
  description = "vpc1"
}

resource "jdcloud_network_security_group" "sg-1" {
  network_security_group_name = "sg-1"
  vpc_id                      = "${jdcloud_vpc.vpc-1.id}"
}

resource "jdcloud_network_security_group_rules" "sg-r-1" {
  network_security_group_id = "${jdcloud_network_security_group.sg-1.id}"

  add_security_group_rules = [{
    address_prefix = "0.0.0.0/0"
    direction      = "0"
    from_port      = "8000"
    protocol       = "300"
    to_port        = "8900"
  }]
}

resource "jdcloud_instance" "vm-1" {
  az            = "cn-north-1a"
  instance_name = "my-vm-1"
  instance_type = "c.n1.large"
  image_id      = "bba85cab-dfdc-4359-9218-7a2de429dd80"
  password      = "${var.vm_password}"
  key_names     = "${jdcloud_key_pairs.key-1.key_name}"
  description   = "Managed by terraform"

  subnet_id              = "${jdcloud_subnet.jd-subnet-1.id}"
  network_interface_name = "veth1"
  primary_ip             = "172.16.0.13"
  secondary_ips          = ["172.16.0.14", "172.16.0.15"]
  # secondary_ip_count   = 2
  security_group_ids     = ["${jdcloud_network_security_group.sg-1.id}"]
  sanity_check           = 1

  elastic_ip_bandwidth_mbps = 10
  elastic_ip_provider       = "bgp"

  system_disk = {
    disk_category = "local"
    auto_delete   = true
    device_name   = "vda"
    no_device     = true
  }

  data_disk = {
    disk_category = "local"
    auto_delete   = true
    device_name   = "vdb"
    no_device     = true
  }

  data_disk = {
    disk_category = "cloud"
    auto_delete   = true
    device_name   = "vdc"
    no_device     = true

    az           = "cn-north-1a"
    disk_name    = "vm1-datadisk-1"
    description  = "test"
    disk_type    = "premium-hdd"
    disk_size_gb = 50
  }
}

resource "jdcloud_subnet" "jd-subnet-1" {
  vpc_id      = "${jdcloud_vpc.vpc-1.id}"
  cidr_block  = "172.16.0.0/26"
  subnet_name = "subnet_example"
  description = "testing"
}

resource "jdcloud_route_table" "jd-route-table-1" {
  vpc_id           = "${jdcloud_vpc.vpc-1.id}"
  route_table_name = "my_route_table_haha"
  description      = "Testing"
}

resource "jdcloud_route_table_association" "route-table-association-1" {
  subnet_id      = "${jdcloud_subnet.jd-subnet-1.id}"
  route_table_id = "${jdcloud_route_table.jd-route-table-1.id}"
}

resource "jdcloud_key_pairs" "keypairs_1" {
  key_name   = "JDCLODU-123312FMK"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAm3c0aR7zI0Xkrm1MD3zDrazC+fR+DV6p/xdzQJWviqPSFGfsatptY3Bc6gYF/qY+Jjccmrm6SKrtW0xPicCw5/uGAuIyhzBnG1Ix0fITdJkeBzyBpxdu/oxnJuvu5P8BLfoFH80ovUqysnttC/7hHBp+uIctkt/g14Kqd7kuPc0Gp4cx7tntNWpmzHJI9i+ayF95nJyFGIjF/s57b9pcKnnv2LXkMDNxsnzgWwPpi2hqGpQSz//ji8GgSED08u34zSjVbPc0TYJy4GO+uD8hozGnf9Evlpqx4OSB0D+4AuRcIniPgCOotYpOdp3Lj7CQRwzkiFZ6YpOxj1qMD4fnjQ== rsa-key-jddevelop"
}

resource "jdcloud_key_pairs" "keypairs_2" {
  key_name = "JDCLODU-123312FMF"
}
