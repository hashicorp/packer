<!-- Code generated from the comments of the Config struct in builder/amazon/ebsvolume/builder.go; DO NOT EDIT MANUALLY -->

-   `ebs_volumes` (BlockDevices) - Add the block device mappings to the AMI. The block device mappings
    allow for keys:
    
-   `ena_support` (*bool) - Enable enhanced networking (ENA but not SriovNetSupport) on
    HVM-compatible AMIs. If set, add ec2:ModifyInstanceAttribute to your AWS
    IAM policy. If false, this will disable enhanced networking in the final
    AMI as opposed to passing the setting through unchanged from the source.
    Note: you must make sure enhanced networking is enabled on your
    instance. See Amazon's documentation on enabling enhanced networking.
    
-   `sriov_support` (bool) - Enable enhanced networking (SriovNetSupport but not ENA) on
    HVM-compatible AMIs. If true, add `ec2:ModifyInstanceAttribute` to your
    AWS IAM policy. Note: you must make sure enhanced networking is enabled
    on your instance. See [Amazon's documentation on enabling enhanced
    networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking).
    Default `false`.
    