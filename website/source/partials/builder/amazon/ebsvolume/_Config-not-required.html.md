<!-- Code generated from the comments of the Config struct in builder/amazon/ebsvolume/builder.go; DO NOT EDIT MANUALLY -->

-   `ebs_volumes` (BlockDevices) - Add the block device mappings to the AMI. If you add instance store
    volumes or EBS volumes in addition to the root device volume, the
    created AMI will contain block device mapping information for those
    volumes. Amazon creates snapshots of the source instance's root volume
    and any other EBS volumes described here. When you launch an instance
    from this new AMI, the instance automatically launches with these
    additional volumes, and will restore them from snapshots taken from the
    source instance. See the [BlockDevices](#block-devices-configuration)
    documentation for fields.
    
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
    