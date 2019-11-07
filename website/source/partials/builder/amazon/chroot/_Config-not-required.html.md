<!-- Code generated from the comments of the Config struct in builder/amazon/chroot/builder.go; DO NOT EDIT MANUALLY -->

-   `ami_block_device_mappings` (awscommon.BlockDevices) - Add one or more [block device
    mappings](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html)
    to the AMI. If this field is populated, and you are building from an
    existing source image, the block device mappings in the source image
    will be overwritten. This means you must have a block device mapping
    entry for your root volume, `root_volume_size` and `root_device_name`.
    See the [BlockDevices](#block-devices-configuration) documentation for
    fields.
    
-   `chroot_mounts` ([][]string) - This is a list of devices to mount into the chroot environment. This
    configuration parameter requires some additional documentation which is
    in the Chroot Mounts section. Please read that section for more
    information on how to use this.
    
-   `command_wrapper` (string) - How to run shell commands. This defaults to {{.Command}}. This may be
    useful to set if you want to set environmental variables or perhaps run
    it with sudo or so on. This is a configuration template where the
    .Command variable is replaced with the command to be run. Defaults to
    {{.Command}}.
    
-   `copy_files` ([]string) - Paths to files on the running EC2 instance that will be copied into the
    chroot environment prior to provisioning. Defaults to /etc/resolv.conf
    so that DNS lookups work. Pass an empty list to skip copying
    /etc/resolv.conf. You may need to do this if you're building an image
    that uses systemd.
    
-   `device_path` (string) - The path to the device where the root volume of the source AMI will be
    attached. This defaults to "" (empty string), which forces Packer to
    find an open device automatically.
    
-   `nvme_device_path` (string) - When we call the mount command (by default mount -o device dir), the
    string provided in nvme_mount_path will replace device in that command.
    When this option is not set, device in that command will be something
    like /dev/sdf1, mirroring the attached device name. This assumption
    works for most instances but will fail with c5 and m5 instances. In
    order to use the chroot builder with c5 and m5 instances, you must
    manually set nvme_device_path and device_path.
    
-   `from_scratch` (bool) - Build a new volume instead of starting from an existing AMI root volume
    snapshot. Default false. If true, source_ami is no longer used and the
    following options become required: ami_virtualization_type,
    pre_mount_commands and root_volume_size. The below options are also
    required in this mode only:
    
-   `mount_options` ([]string) - Options to supply the mount command when mounting devices. Each option
    will be prefixed with -o and supplied to the mount command ran by
    Packer. Because this command is ran in a shell, user discretion is
    advised. See this manual page for the mount command for valid file
    system specific options.
    
-   `mount_partition` (string) - The partition number containing the / partition. By default this is the
    first partition of the volume, (for example, xvda1) but you can
    designate the entire block device by setting "mount_partition": "0" in
    your config, which will mount xvda instead.
    
-   `mount_path` (string) - The path where the volume will be mounted. This is where the chroot
    environment will be. This defaults to
    /mnt/packer-amazon-chroot-volumes/{{.Device}}. This is a configuration
    template where the .Device variable is replaced with the name of the
    device where the volume is attached.
    
-   `post_mount_commands` ([]string) - As pre_mount_commands, but the commands are executed after mounting the
    root device and before the extra mount and copy steps. The device and
    mount path are provided by {{.Device}} and {{.MountPath}}.
    
-   `pre_mount_commands` ([]string) - A series of commands to execute after attaching the root volume and
    before mounting the chroot. This is not required unless using
    from_scratch. If so, this should include any partitioning and filesystem
    creation commands. The path to the device is provided by {{.Device}}.
    
-   `root_device_name` (string) - The root device name. For example, xvda.
    
-   `root_volume_size` (int64) - The size of the root volume in GB for the chroot environment and the
    resulting AMI. Default size is the snapshot size of the source_ami
    unless from_scratch is true, in which case this field must be defined.
    
-   `root_volume_type` (string) - The type of EBS volume for the chroot environment and resulting AMI. The
    default value is the type of the source_ami, unless from_scratch is
    true, in which case the default value is gp2. You can only specify io1
    if building based on top of a source_ami which is also io1.
    
-   `source_ami_filter` (awscommon.AmiFilterOptions) - Filters used to populate the source_ami field. Example:
    
        ``` json
        {
          "source_ami_filter": {
          "filters": {
           "virtualization-type": "hvm",
           "name": "ubuntu/images/*ubuntu-xenial-16.04-amd64-server-*",
           "root-device-type": "ebs"
         },
         "owners": ["099720109477"],
         "most_recent": true
          }
        }
        ```
    
        This selects the most recent Ubuntu 16.04 HVM EBS AMI from Canonical. NOTE:
        This will fail unless *exactly* one AMI is returned. In the above example,
        `most_recent` will cause this to succeed by selecting the newest image.
    
        -   `filters` (map of strings) - filters used to select a `source_ami`.
            NOTE: This will fail unless *exactly* one AMI is returned. Any filter
            described in the docs for
            [DescribeImages](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeImages.html)
            is valid.
    
        -   `owners` (array of strings) - Filters the images by their owner. You
            may specify one or more AWS account IDs, "self" (which will use the
            account whose credentials you are using to run Packer), or an AWS owner
            alias: for example, "amazon", "aws-marketplace", or "microsoft". This
            option is required for security reasons.
    
        -   `most_recent` (boolean) - Selects the newest created image when true.
            This is most useful for selecting a daily distro build.
    
        You may set this in place of `source_ami` or in conjunction with it. If you
        set this in conjunction with `source_ami`, the `source_ami` will be added
        to the filter. The provided `source_ami` must meet all of the filtering
        criteria provided in `source_ami_filter`; this pins the AMI returned by the
        filter, but will cause Packer to fail if the `source_ami` does not exist.
    
-   `root_volume_tags` (awscommon.TagMap) - Tags to apply to the volumes that are *launched*. This is a [template
    engine](/docs/templates/engine.html), see [Build template
    data](#build-template-data) for more information.
    
-   `ami_architecture` (string) - what architecture to use when registering the final AMI; valid options
    are "x86_64" or "arm64". Defaults to "x86_64".
    