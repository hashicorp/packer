-   `delete_on_termination` (boolean) - Indicates whether the EBS volume is
    deleted on instance termination. Default `false`. **NOTE**: If this
    value is not explicitly set to `true` and volumes are not cleaned up by
    an alternative method, additional volumes will accumulate after
    every build.

-   `device_name` (string) - The device name exposed to the instance (for
     example, "/dev/sdh" or "xvdh"). Required when specifying `volume_size`.

-   `encrypted` (boolean) - Indicates whether to encrypt the volume or not

-   `iops` (integer) - The number of I/O operations per second (IOPS) that the
    volume supports. See the documentation on
    [IOPs](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_EbsBlockDevice.html)
    for more information

-   `no_device` (boolean) - Suppresses the specified device included in the
    block device mapping of the AMI

-   `snapshot_id` (string) - The ID of the snapshot

-   `virtual_name` (string) - The virtual device name. See the documentation on
    [Block Device
    Mapping](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_BlockDeviceMapping.html)
    for more information

-   `volume_size` (integer) - The size of the volume, in GiB. Required if not
    specifying a `snapshot_id`

-   `volume_type` (string) - The volume type. gp2 for General Purpose (SSD)
    volumes, io1 for Provisioned IOPS (SSD) volumes, and standard for Magnetic
    volumes

