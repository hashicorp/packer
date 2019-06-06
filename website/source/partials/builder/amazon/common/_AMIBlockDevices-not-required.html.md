<!-- Code generated from the comments of the AMIBlockDevices struct in builder/amazon/common/block_device.go; DO NOT EDIT MANUALLY -->

-   `ami_block_device_mappings` ([]BlockDevice) - Add one or more [block device
    mappings](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html)
    to the AMI. These will be attached when booting a new instance from your
    AMI. To add a block device during the Packer build see
    `launch_block_device_mappings` below. Your options here may vary
    depending on the type of VM you use. The block device mappings allow for
    the following configuration:
    -   `delete_on_termination` (boolean) - Indicates whether the EBS volume is
         deleted on instance termination. Default `false`. **NOTE**: If this
         value is not explicitly set to `true` and volumes are not cleaned up by
         an alternative method, additional volumes will accumulate after every
         build.
    
    -   `device_name` (string) - The device name exposed to the instance (for
        example, `/dev/sdh` or `xvdh`). Required for every device in the block
        device mapping.
    
    -   `encrypted` (boolean) - Indicates whether or not to encrypt the volume.
        By default, Packer will keep the encryption setting to what it was in
        the source image. Setting `false` will result in an unencrypted device,
        and `true` will result in an encrypted one.
    
    -   `iops` (number) - The number of I/O operations per second (IOPS) that
        the volume supports. See the documentation on
        [IOPs](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_EbsBlockDevice.html)
        for more information
    
    -   `kms_key_id` (string) - The ARN for the KMS encryption key. When
         specifying `kms_key_id`, `encrypted` needs to be set to `true`. For
         valid formats see *KmsKeyId* in the [AWS API docs -
         CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html).
    
    -   `no_device` (boolean) - Suppresses the specified device included in the
        block device mapping of the AMI.
    
    -   `snapshot_id` (string) - The ID of the snapshot.
    
    -   `virtual_name` (string) - The virtual device name. See the
        documentation on [Block Device
        Mapping](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_BlockDeviceMapping.html)
        for more information.
    
    -   `volume_size` (number) - The size of the volume, in GiB. Required if
        not specifying a `snapshot_id`.
    
    -   `volume_type` (string) - The volume type. `gp2` for General Purpose
        (SSD) volumes, `io1` for Provisioned IOPS (SSD) volumes, `st1` for
        Throughput Optimized HDD, `sc1` for Cold HDD, and `standard` for
        Magnetic volumes.
    