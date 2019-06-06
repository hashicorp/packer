<!-- Code generated from the comments of the LaunchBlockDevices struct in builder/amazon/common/block_device.go; DO NOT EDIT MANUALLY -->

-   `launch_block_device_mappings` ([]BlockDevice) - Add one or more block devices before the Packer build starts. If you add
    instance store volumes or EBS volumes in addition to the root device
    volume, the created AMI will contain block device mapping information
    for those volumes. Amazon creates snapshots of the source instance's
    root volume and any other EBS volumes described here. When you launch an
    instance from this new AMI, the instance automatically launches with
    these additional volumes, and will restore them from snapshots taken
    from the source instance.
    
    In addition to the fields available in ami_block_device_mappings, you
    may optionally use the following field:
    -   `omit_from_artifact` (boolean) - If true, this block device will not
        be snapshotted and the created AMI will not contain block device mapping
        information for this volume. If false, the block device will be mapped
        into the final created AMI. Set this option to true if you need a block
        device mounted in the surrogate AMI but not in the final created AMI.
    