# Motivation 

## Why do I need this "Scanner" ?
Consider a a case when you would like to deploy some destructive test. 
In order to re-build your infrastructure in the quickest way possible, Scanner scan your infrastructure and generate a file to back up your current status. 
By rebuilding your infrastructure,  you just need to command terraform apply with this back up file. *(easy-peasy.....)*

## Why not just copy the current configuration file ?
You can.   

**However**, most of customers begin their journey with Terraform when they already have some existing resources.  
As is discussed in [issue#15608](https://github.com/hashicorp/terraform/issues/15608). 
Currently , Terraform can not take over existing resources while generating a back up file. 
This means even though you have imported some existing resources, 
you probably still need to full fill your configuration file manually.
For a large scale of infrastructure , it does not seems easy. 

To implement a complete, fully tested "Scanner" is not easy, since a complex dependency among resources is included.  
But at least I could implement a naive one. Scripts on scanner is pretty easy and naive, if you have any furthur plan and suggestions, just leave it :-)

# How to use ?

## 1.Export keys to environment

```bash
export access_key = <Your_access_key>
export secret_key = <Your_secret_key>
export region = <Your_region>
```

## 2. Launch 

```makefile
make scan
```

After Scanner finished, infrastructures will be saved in `scanner/jdcloud.tf`
