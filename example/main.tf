provider "jdcloud" {
  access_key = "${var.access_key}"
  secret_key = "${var.secret_key}"
  region = "${var.region}"
}

# -------------------------------------------------------------DISK

################################################
# 1. Create a disk
################################################
# You can choose to attach a disk to an instance after creating it.
# Alternatively, you can create a disk while creating an instance
# -> See it instance part
# [WARN] If (charge_mode == prepaid_by_duration)
# You can not delete it before they expired. "postpaid_by_usage" is recommended
resource "jdcloud_disk" "disk_test_1" {
  az           = "cn-north-1a"
  name         = "test_disk"
  disk_type    = "premium-hdd"
  disk_size_gb = 60
  charge_mode  = "postpaid_by_usage"
}

################################################
#  2. Attach a disk to an instance
################################################
# There are two ways to specify an resource, you can either
# specify its ID number, like following:
resource "jdcloud_disk_attachment" "disk-attachment-TEST-1"{
  instance_id = "i-exampleid"
  disk_id = "vol-exampleid"
  auto_delete = false
}

# Or you can reference it with its resource name, like following:
resource "jdcloud_disk_attachment" "disk-attachment-TEST-2"{
  instance_id = "${jdcloud_instance.vm-1.id}"
  disk_id = "${jdcloud_disk.disk_test_1.id}"
  auto_delete = false
}

# You'll probably interested in REQUIRED and OPTIONAL parameters
# See in -> github.com/jdclouddevelopers/terraform-provider-jdcloud/blob/master/website/docs


# ------------------------------------------------------------- VPC
resource "jdcloud_vpc" "jd-vpc-1" {
  vpc_name = "my-vpc-1"
  cidr_block = "172.16.0.0/20"
  description = "vpc1"
}

# ------------------------------------------------------------- SUBNET
# CIDR-Block of subnet must be a subset of VPC it belongs to
# e.g : [jd-subnet-1]   <     [jd-vpc-1]    ,thus we have :
#      [172.16.0.0/26]  <  [172.16.0.0/20]
# Besides,there is no overlap among subnet CIDR under same VPC
resource "jdcloud_subnet" "jd-subnet-1" {
  vpc_id = "${jdcloud_vpc.jd-vpc-1.id}"
  cidr_block = "172.16.0.0/26"
  subnet_name = "subnet_example"
  description = "testing"
}

# ---------------------------------------------------------- ROUTE-TABLE
################################################
# 1. Create a Route Table
################################################
resource "jdcloud_route_table" "jd-route-table-1" {
  vpc_id = "${jdcloud_vpc.jd-vpc-1.id}"
  route_table_name = "example_route_table"
  description = "DevOps2018"
}

################################################
# 2. Associate a Route Table to a Subnet
################################################
resource "jdcloud_route_table_association" "rt-association-1" {
  subnet_id = "${jdcloud_subnet.jd-subnet-1.id}"
  route_table_id = "${jdcloud_route_table.jd-route-table-1.id}"
}

################################################
# 3. Add rules to this Route Table
################################################
# Candidates for "next_hop_type" : instance/internet/vpc_peering/bgw
# "address_prefix" : if (next_hop_type == "internet") then there's
#  no overlap between address prefixes. Default priority for a rule is 100
resource "jdcloud_route_table_rules" "rule-example"{
  route_table_id = "rtb-example"
  rule_specs = [{
    next_hop_type = "internet"
    next_hop_id   = "internet"
    address_prefix= "10.0.0.0/16"
    priority      = 100
  }]
}

# ---------------------------------------------------------- SECURITY-GROUP
################################################
# 1. Create a Security Group
################################################
resource "jdcloud_network_security_group" "sg-1" {
  network_security_group_name = "sg-1"
  vpc_id = "${jdcloud_vpc.jd-vpc-1.id}"
}

################################################
# 2. Create Security Group Rules
################################################
# Unlike route table rules, SG rules are defined under same resource.
# Detailed candidates, together with their explaination, Can be found here:
# github.com/jdclouddevelopers/terraform-provider-jdcloud/blob/master/website/docs/jdcloud_network_security_group_rules.html.markdown
resource "jdcloud_network_security_group_rules" "sg-r-1" {
  security_group_id = "${jdcloud_network_security_group.sg-1.id}"
  security_group_rules = [
    {
      address_prefix = "0.0.0.0/0"
      direction = "0"
      from_port = "8000"
      protocol = "300"
      to_port = "8900"
    },
    {
      address_prefix = "0.0.0.0/0"
      direction = "1"
      from_port = "8000"
      protocol = "300"
      to_port = "8900"
    },
  ]
}

# ---------------------------------------------------------- NETWORK-INTERFACE
################################################
# 1. Create a Network Interface
################################################
# It can be confusing on parameter [secondary_ip_addresses] & [secondary_ip_count]
# The first one represents you would like to specify some address for this network
# interface, while the second one represents you want more addresses, but not caring what they actually are
# For example, for a following config, you'll get in total 3 addresses on this NI
# The first one is 10.0.3.0, and the remaining two may be 10.0.4.0 and 10.0.3.1
resource "jdcloud_network_interface" "ineterface-TEST-1" {
  subnet_id = "subnet-example"
  description = "test"
  az = "cn-north-1"
  network_interface_name = "TerraformTest"
  secondary_ip_addresses = ["10.0.3.0",]
  secondary_ip_count = "2"
  security_groups = ["sg-example"]
}

################################################
# 2. Associate a Network Interface to Instance
################################################
# You can create a NI and attach it to an Instance, Alternatively
# You can still create a NI while creating an Instance -> See instance part
resource "jdcloud_network_interface_attachment" "attachment-TEST-1"{
  instance_id = "i-example"
  network_interface_id = "port-example"
  auto_delete = "true"
}

# ---------------------------------------------------------- ELASTIC-IP
################################################
# 1. Create a Elastic-IP
################################################
# "eip_provider" = bgp/no_bgp, selected according to your region
# cn-north-1：bgp；cn-south-1：[bgp，no_bgp]；cn-east-1：[bgp，no_bgp]；cn-east-2：bgp
resource "jdcloud_eip" "eip-TEST-1"{
  eip_provider = "bgp"
  bandwidth_mbps = 1
}

################################################
# 2. Associate an EIP with an Instance
################################################
# Similarly, you can create -> Associate
# Or creating IP while creating instance
resource "jdcloud_eip_association" "eip-association-TEST-1"{
  instance_id = "i-p3yh27xd3s"
  elastic_ip_id = "fip-e3lfigpewx"
}

# ---------------------------------------------------------- OSS-Bucket
################################################
# 1. Create a OSS-Bucket with certain ACL
################################################
# ACL means privilege control to this bucket
# PRIVATE : only you can write
# public-read: Owner has full control, other people can read from this but no writing is allowed
# public-read-write: Everyone can read/write from this bucket
resource "jdcloud_oss_bucket" "jd-bucket-2" {
  bucket_name = "example"
  acl = "private"
}

################################################
# 2. Upload file to this bucket
################################################
# When you would like to upload some files to
# to a bucket, buckets are specified by it name
resource "jdcloud_oss_bucket_upload" "devops" {
  bucket_name = "example"
  file_name = "/home/DevOps/hello.txt"
}


# ---------------------------------------------------------- Instance
# As we can see that although it seems to be pretty long with loads of
# parameter to specify. Most of them are component attaching,NI,EIP,Disk,etc..
# Values in these fields has no difference with above
# [WARN]:
resource "jdcloud_instance" "vm-1" {

  ################################################
  # 1. VM Config
  ################################################
  # Password is a optional field. By missing this
  # Field,passwords will be sent by email and SMS
  az = "cn-north-1a"
  instance_name = "my-vm-1"
  instance_type = "c.n1.large"
  image_id = "example_image_id"
  password = "${var.vm_password}"
  key_names = "${jdcloud_key_pairs.key-1.key_name}"
  description = "Managed by terraform"

  ################################################
  # 2. Create a Network Interface with it
  ################################################
  subnet_id = "${jdcloud_subnet.jd-subnet-1.id}"
  network_interface_name = "veth1"
  sanity_check = 1
  primary_ip = "172.16.0.13"
  secondary_ips = [
    "172.16.0.14",
    "172.16.0.15"]
  security_group_ids = ["${jdcloud_network_security_group.sg-1.id}"]

  ################################################
  # 3. Create an EIP with it
  ################################################
  elastic_ip_bandwidth_mbps = 10
  elastic_ip_provider = "bgp"

  ################################################
  # 4. System-Disk(Required)
  ################################################
  # We would recommend local disk as system disk :
  # System-Disk ─┬── "disk_category" = "local" ──> Always work ,disk size fixed to 40Gb
  #             └   "disk_category" = "cloud" ──> Works only when az == cn-east
  system_disk = {
    disk_category = "local"
    auto_delete = true
    device_name = "vda"
  }

  ################################################
  # 5. Data-Disk(Optional)
  ################################################
  # You can attach multiple data-disk with this instance
  # Device name for disk must be unique
  data_disk = [
  {
    disk_category = "local"
    auto_delete = true
    device_name = "vdb"
  },
  {
    disk_category = "cloud"
    auto_delete = true
    device_name = "vdc"

    az = "cn-north-1a"
    disk_name = "vm1-datadisk-1"
    description = "test"
    disk_type = "ssd"
    disk_size_gb = 50
  }]
}


# ---------------------------------------------------------- RDS
################################################
# 1. Create an Instance
################################################
# [WARN] If (charge_mode == prepaid_by_duration)
# You can not delete it before they expired. "postpaid_by_duration" is recommended
resource "jdcloud_rds_instance" "rds-test"{
  instance_name = "test"
  engine = "MySQL"
  engine_version = "5.7"
  instance_class = "db.mysql.s1.micro"
  instance_storage_gb = "20"
  az = "cn-north-1a"
  vpc_id = "vpc-example"
  subnet_id = "subnet-example"
  charge_mode = "postpaid_by_duration"
}

################################################
# 1. Create accounts on this instance
################################################
resource "jdcloud_rds_account" "rds-test1"{
  instance_id = "mysql-example"
  username = "DevOps"
  password = "JDCloud123"
}

################################################
# 2. Create databases on this accounts
################################################
# [WARN] Currently any modification on Database resource
# is banned. Trying to modify will result in returning errors
resource "jdcloud_rds_database" "db-TEST"{
  instance_id = "mysql-g0afoqpl6y"
  db_name = "cloudb1"
  character_set = "utf8"
}
resource "jdcloud_rds_database" "db-TEST-2"{
  instance_id = "mysql-g0afoqpl6y"
  db_name = "cloudb2"
  character_set = "utf8"
}

################################################
# 3. Grant privilege for user accounts
################################################
resource "jdcloud_rds_privilege" "pri-test" {
  instance_id = "mysql-g0afoqpl6y"
  username = "DevOps"
  account_privilege = [
    {db_name = "cloudb1",privilege = "ro"},
    {db_name = "cloudb2",privilege = "rw"},
  ]
}

# ---------------------------------------------------------- INSTANCE-TEMPLATE
####################################################################################################
# Full parameters ->
#
#   + instance_type : g.n2.medium/g.n2.large...Just different type of instance , by default its g.n2.medium
#   + password : Optional, if you leave it blank. password will be sent to you by email and SMS.
#   + image_id : If you would like to start your instance from an image , fill in here, by default its Ubuntu:16.04
#   + bandwidth: Optional. if you leave it blank, no public IP will be assigned to this instance
#   + ip_service_provider : [BGP/nonBGP] , by default it's set to BGP
#   + charge_mode : [bandwith/flow] , by default it's set to bandwith
#   + subnet_id : Each instance is supposed to exists under a subnet, fill its id here
#   + security_group_ids : You can make your instance exists under multuple SGs, fill them here as an array
#   + local_disk/data_disk
#       - disk_category : [local/cloud] : For system disk, local disk is recommended
#       - disk_type : [premium-hdd/ssd] : For system disk, we would recommend premium-hdd
#                                         For data disk , ssd, since premium-hdd ususally out of stock
#       - auto_delete: [true/false] : Disks will be deleted automatically once it's been detached from instance
#       - disk_size : [20~1000] - Must be multiples of 10, that's 20,30,40... by default it is set  to 40GB
#       - device_name : [vda/vdb/vdc....] - Attach point of disk, points must be unique among all disks
#                                       we would recommend to leave it blank if you're not farmiliar with it
#       - snapshot_id : If you would like to build a template from snapshot, fill in its id here
####################################################################################################
resource "jdcloud_instance_template" "instance_template" {
  "template_name" = "<Name it as you like>"
  "password" = "<Give it a password>"
  "instance_type" = "g.n2.medium"
  "image_id" = "img-example"
  "bandwidth" = 5
  "description" = "GrandJDcloud"
  "ip_service_provider" = "BGP"
  "charge_mode" = "bandwith"
  "subnet_id" = "subnet-exmaple"
  "security_group_ids" = [" sg-example"]
  "system_disk" = {
    "disk_category" = "local"
    "disk_type" = ""
  }
  "data_disks" = {
    disk_category = "cloud"
  }

}
# ---------------------------------------------------------- AVAILABILITY-GROUP
# Parameters and candidates
#     1. [AZ] is an array which, candidates lists as follows ["cn-north-1a","cn-north-1b","cn-east-1b","cn-east-1a","cn-south-1a"]
#     2. [AG_TYPE] is a string, candidates are : "kvm", "docker"
# [Ag_name] and [Description] are updatable while the remaining not.
#
resource "jdcloud_availability_group" "ag_01" {
  availability_group_name = "example_ag_name"
  az = ["cn-north-1a","cn-north-1b"]
  instance_template_id = "example_template_id"
  description  = "This is an example description"
  ag_type = "docker"
}

# ---------------------------------------------------------- KEY-PAIRS

resource "jdcloud_key_pairs" "key-1" {
  key_name = "JDCLODU-123312FMK"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAm3c0aR7zI0Xkrm1MD3zDrazC+fR+DV6p/xdzQJWviqPSFGfsatptY3Bc6gYF/qY+Jjccmrm6SKrtW0xPicCw5/uGAuIyhzBnG1Ix0fITdJkeBzyBpxdu/oxnJuvu5P8BLfoFH80ovUqysnttC/7hHBp+uIctkt/g14Kqd7kuPc0Gp4cx7tntNWpmzHJI9i+ayF95nJyFGIjF/s57b9pcKnnv2LXkMDNxsnzgWwPpi2hqGpQSz//ji8GgSED08u34zSjVbPc0TYJy4GO+uD8hozGnf9Evlpqx4OSB0D+4AuRcIniPgCOotYpOdp3Lj7CQRwzkiFZ6YpOxj1qMD4fnjQ== rsa-key-jddevelop"
}

resource "jdcloud_key_pairs" "keypairs_2" {
  key_name = "JDCLODU-123312FMF"
  key_file = "private.ppk"
}