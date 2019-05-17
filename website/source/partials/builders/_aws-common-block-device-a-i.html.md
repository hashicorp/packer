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