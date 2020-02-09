<!-- Code generated from the comments of the BlockDevice struct in builder/amazon/common/block_device.go; DO NOT EDIT MANUALLY -->

-   `delete_on_termination` (bool) - Indicates whether the EBS volume is deleted on instance termination.
    Default false. NOTE: If this value is not explicitly set to true and
    volumes are not cleaned up by an alternative method, additional volumes
    will accumulate after every build.
    
-   `device_name` (string) - The device name exposed to the instance (for example, /dev/sdh or xvdh).
    Required for every device in the block device mapping.
    
-   `encrypted` (boolean) - Indicates whether or not to encrypt the volume. By default, Packer will
    keep the encryption setting to what it was in the source image. Setting
    false will result in an unencrypted device, and true will result in an
    encrypted one.
    
-   `iops` (int64) - The number of I/O operations per second (IOPS) that the volume supports.
    See the documentation on
    [IOPs](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_EbsBlockDevice.html)
    for more information
    
-   `no_device` (bool) - Suppresses the specified device included in the block device mapping of
    the AMI.
    
-   `snapshot_id` (string) - The ID of the snapshot.
    
-   `virtual_name` (string) - The virtual device name. See the documentation on Block Device Mapping
    for more information.
    
-   `volume_type` (string) - The volume type. gp2 for General Purpose (SSD) volumes, io1 for
    Provisioned IOPS (SSD) volumes, st1 for Throughput Optimized HDD, sc1
    for Cold HDD, and standard for Magnetic volumes.
    
-   `volume_size` (int64) - The size of the volume, in GiB. Required if not specifying a
    snapshot_id.
    
-   `kms_key_id` (string) - ID, alias or ARN of the KMS key to use for boot volume encryption. This
    only applies to the main region, other regions where the AMI will be
    copied will be encrypted by the default EBS KMS key. For valid formats
    see KmsKeyId in the [AWS API docs -
    CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html)
    This field is validated by Packer, when using an alias, you will have to
    prefix kms_key_id with alias/.
    