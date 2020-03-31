<!-- Code generated from the comments of the BlockDevice struct in builder/amazon/common/block_device.go; DO NOT EDIT MANUALLY -->
These will be attached when booting a new instance from your AMI. Your
options here may vary depending on the type of VM you use.

Example use case:

The following mapping will tell Packer to encrypt the root volume of the
build instance at launch using a specific non-default kms key:

```json
[{
		"device_name": "/dev/sda1",
		"encrypted": true,
		"kms_key_id": "1a2b3c4d-5e6f-1a2b-3c4d-5e6f1a2b3c4d"
}]
```

Documentation for Block Devices Mappings can be found here:
https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html
