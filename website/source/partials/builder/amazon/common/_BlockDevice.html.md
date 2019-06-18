<!-- Code generated from the comments of the BlockDevice struct in builder/amazon/common/block_device.go; DO NOT EDIT MANUALLY -->
These will be attached when booting a new instance from your AMI.
Your options here may vary depending on the type of VM you use. Example:

``` json
"builders":[{
"type":"...",
"ami_block_device_mappings":[{
          "device_name":"xvda",
          "delete_on_termination":true,
          "volume_type":"gp2"
    }]
 }
```
Documentation for Block Devices Mappings can be found here:
https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html
