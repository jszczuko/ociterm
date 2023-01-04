


# Disclamer

This project is NOT Oralce product. It's just application I've written in free time as functionality demonstrator. 

**All bugs are mine and mine alone.**

# TLDR;

# Table of contents
1. [Motivation](#motivation)
2. [Configuration](#configuration)

# Motivation <a name="motivation"></a>

[OCI](https://www.oracle.com/cloud/) provides awesome tools to manage resources. Both [OCI console](https://www.oracle.com/cloud/sign-in.html) and [OCI CLI](https://github.com/oracle/oci-cli) are more than enough to mange resources in any way necessary.  But hammer is perfect tool when you have to hammer a nail, when you have to paint a wall it tends to be less useful. 
I found myself in specific use case condition. I have to switch between [tenancies](https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/managingtenancy.htm) very rapidly and I normally operate with name of resource as key value not [OCID](https://docs.oracle.com/en-us/iaas/Content/General/Concepts/identifiers.htm).  Very often action required is very simple, like START/STOP/RESTART compute instance and I found that most time in these cases is used for moving between tenancies and it's basically wasted. 

To switch between tenancies in OCI console you need to logout and login to new tenancy and if you are doing this often enough depending on browser you are using, you can run into a problem of cached sessions. 

With OCI CLI it's better as tenancy is part of [profile in configuration file header](https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliconfigure.htm), but you have to know OCID of resource that you want to manage. 

Therefore I needed a simple tool to allow me to:
- switch between tenancies as fast as possible using tenancy name;
- operate on resource name rather than OCID;
- execute basıc operations (for starters on compute instance).
	
In addition I wanted to work on some useful tool in [GOLANG](https://go.dev/). 

# Configuration <a name="configuration"></a>

Application is suing subset of standard [OCI CLI configuration file](https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliconfigure.htm) located in by default in ```$HOME/.oci/config```.
Minimum required information is described below:

```properties
[tenancy_dev]
user=ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
fingerprint=aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa
key_file=/Users/jszczuko/.oci/oci_api_key.pem
tenancy=ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
region=us-ashburn-1

[tenancy_qa]
user=ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
fingerprint=aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa:aa
key_file=/Users/jszczuko/.oci/oci_api_key.pem
tenancy=ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
region=us-ashburn-1
```

Headers of profiles ```[profile_name]``` will be used as key values for application.