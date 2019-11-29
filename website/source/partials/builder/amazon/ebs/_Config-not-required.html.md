<!-- Code generated from the comments of the Config struct in builder/amazon/ebs/builder.go; DO NOT EDIT MANUALLY -->

-   `ami_block_device_mappings` (awscommon.BlockDevices) - Add one or more block device mappings to the AMI. These will be attached
    when booting a new instance from your AMI. To add a block device during
    the Packer build see `launch_block_device_mappings` below. Your options
    here may vary depending on the type of VM you use. See the
    [BlockDevices](#block-devices-configuration) documentation for fields.
    
-   `launch_block_device_mappings` (awscommon.BlockDevices) - Add one or more block devices before the Packer build starts. If you add
    instance store volumes or EBS volumes in addition to the root device
    volume, the created AMI will contain block device mapping information
    for those volumes. Amazon creates snapshots of the source instance's
    root volume and any other EBS volumes described here. When you launch an
    instance from this new AMI, the instance automatically launches with
    these additional volumes, and will restore them from snapshots taken
    from the source instance. See the
    [BlockDevices](#block-devices-configuration) documentation for fields.
    
-   `run_volume_tags` (awscommon.TagMap) - Tags to apply to the volumes that are *launched* to create the AMI.
    These tags are *not* applied to the resulting AMI unless they're
    duplicated in `tags`. This is a [template
    engine](/docs/templates/engine.html), see [Build template
    data](#build-template-data) for more information.
    
-   `no_ephemeral` (bool) - Relevant only to Windows guests: If you set this flag, we'll add clauses
    to the launch_block_device_mappings that make sure ephemeral drives
    don't show up in the EC2 console. If you launched from the EC2 console,
    you'd get this automatically, but the SDK does not provide this service.
    For more information, see
    https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/InstanceStorage.html.
    Because we don't validate the OS type of your guest, it is up to you to
    make sure you don't set this for *nix guests; behavior may be
    unpredictable.
    