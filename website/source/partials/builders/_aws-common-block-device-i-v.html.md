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