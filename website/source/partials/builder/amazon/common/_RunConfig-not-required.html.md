<!-- Code generated from the comments of the RunConfig struct in builder/amazon/common/run_config.go; DO NOT EDIT MANUALLY -->

-   `associate_public_ip_address` (bool) - If using a non-default VPC,
    public IP addresses are not provided by default. If this is true, your
    new instance will get a Public IP. default: false
    
-   `availability_zone` (string) - Destination availability zone to launch
    instance in. Leave this empty to allow Amazon to auto-assign.
    
-   `block_duration_minutes` (int64) - Requires spot_price to be set. The
    required duration for the Spot Instances (also known as Spot blocks). This
    value must be a multiple of 60 (60, 120, 180, 240, 300, or 360). You can't
    specify an Availability Zone group or a launch group if you specify a
    duration.
    
-   `disable_stop_instance` (bool) - Packer normally stops the build instance after all provisioners have
    run. For Windows instances, it is sometimes desirable to [run
    Sysprep](http://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/ami-create-standard.html)
    which will stop the instance for you. If this is set to `true`, Packer
    *will not* stop the instance but will assume that you will send the stop
    signal yourself through your final provisioner. You can do this with a
    [windows-shell
    provisioner](https://www.packer.io/docs/provisioners/windows-shell.html).
    Note that Packer will still wait for the instance to be stopped, and
    failing to send the stop signal yourself, when you have set this flag to
    `true`, will cause a timeout.
    Example of a valid shutdown command:
    
    ``` json
    {
      "type": "windows-shell",
      "inline": ["\"c:\\Program Files\\Amazon\\Ec2ConfigService\\ec2config.exe\" -sysprep"]
    }
    ```
    
-   `ebs_optimized` (bool) - Mark instance as [EBS
    Optimized](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html).
    Default `false`.
    
-   `enable_t2_unlimited` (bool) - Enabling T2 Unlimited allows the source instance to burst additional CPU
    beyond its available [CPU
    Credits](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/t2-credits-baseline-concepts.html)
    for as long as the demand exists. This is in contrast to the standard
    configuration that only allows an instance to consume up to its
    available CPU Credits. See the AWS documentation for [T2
    Unlimited](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/t2-unlimited.html)
    and the **T2 Unlimited Pricing** section of the [Amazon EC2 On-Demand
    Pricing](https://aws.amazon.com/ec2/pricing/on-demand/) document for
    more information. By default this option is disabled and Packer will set
    up a [T2
    Standard](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/t2-std.html)
    instance instead.
    
    To use T2 Unlimited you must use a T2 instance type, e.g. `t2.micro`.
    Additionally, T2 Unlimited cannot be used in conjunction with Spot
    Instances, e.g. when the `spot_price` option has been configured.
    Attempting to do so will cause an error.
    
    !&gt; **Warning!** Additional costs may be incurred by enabling T2
    Unlimited - even for instances that would usually qualify for the
    [AWS Free Tier](https://aws.amazon.com/free/).
    
-   `iam_instance_profile` (string) - The name of an [IAM instance
    profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/instance-profiles.html)
    to launch the EC2 instance with.
    
-   `temporary_iam_instance_profile_policy_document` (\*PolicyDocument) - Temporary IAM instance profile policy document
    If IamInstanceProfile is specified it will be used instead. Example:
    
    ```json
    {
    	"Version": "2012-10-17",
    	"Statement": [
    		{
    			"Action": [
    			"logs:*"
    			],
    			"Effect": "Allow",
    			"Resource": "*"
    		}
    	]
    }
    ```
    
-   `shutdown_behavior` (string) - Automatically terminate instances on
    shutdown in case Packer exits ungracefully. Possible values are stop and
    terminate. Defaults to stop.
    
-   `security_group_filter` (SecurityGroupFilterOptions) - Filters used to populate the `security_group_ids` field. Example:
    
    ``` json
    {
      "security_group_filter": {
        "filters": {
          "tag:Class": "packer"
        }
      }
    }
    ```
    
    This selects the SG's with tag `Class` with the value `packer`.
    
    -   `filters` (map of strings) - filters used to select a
        `security_group_ids`. Any filter described in the docs for
        [DescribeSecurityGroups](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroups.html)
        is valid.
    
    `security_group_ids` take precedence over this.
    
-   `run_tags` (map[string]string) - Tags to apply to the instance that is that is *launched* to create the
    EBS volumes. This is a [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.
    
-   `security_group_id` (string) - The ID (not the name) of the security
    group to assign to the instance. By default this is not set and Packer will
    automatically create a new temporary security group to allow SSH access.
    Note that if this is specified, you must be sure the security group allows
    access to the ssh_port given below.
    
-   `security_group_ids` ([]string) - A list of security groups as
    described above. Note that if this is specified, you must omit the
    security_group_id.
    
-   `source_ami_filter` (AmiFilterOptions) - Filters used to populate the `source_ami`
    field. Example:
    
      ``` json
      {
        "source_ami_filter": {
          "filters": {
            "virtualization-type": "hvm",
            "name": "ubuntu/images/\*ubuntu-xenial-16.04-amd64-server-\*",
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
          alias: for example, `amazon`, `aws-marketplace`, or `microsoft`. This
          option is required for security reasons.
    
      -   `most_recent` (boolean) - Selects the newest created image when true.
          This is most useful for selecting a daily distro build.
    
      You may set this in place of `source_ami` or in conjunction with it. If you
      set this in conjunction with `source_ami`, the `source_ami` will be added
      to the filter. The provided `source_ami` must meet all of the filtering
      criteria provided in `source_ami_filter`; this pins the AMI returned by the
      filter, but will cause Packer to fail if the `source_ami` does not exist.
    
-   `spot_instance_types` ([]string) - a list of acceptable instance
    types to run your build on. We will request a spot instance using the max
    price of spot_price and the allocation strategy of "lowest price".
    Your instance will be launched on an instance type of the lowest available
    price that you have in your list.  This is used in place of instance_type.
    You may only set either spot_instance_types or instance_type, not both.
    This feature exists to help prevent situations where a Packer build fails
    because a particular availability zone does not have capacity for the
    specific instance_type requested in instance_type.
    
-   `spot_price` (string) - The maximum hourly price to pay for a spot instance
    to create the AMI. Spot instances are a type of instance that EC2 starts
    when the current spot price is less than the maximum price you specify.
    Spot price will be updated based on available spot instance capacity and
    current spot instance requests. It may save you some costs. You can set
    this to auto for Packer to automatically discover the best spot price or
    to "0" to use an on demand instance (default).
    
-   `spot_price_auto_product` (string) - Required if spot_price is set to
    auto. This tells Packer what sort of AMI you're launching to find the
    best spot price. This must be one of: Linux/UNIX, SUSE Linux,
    Windows, Linux/UNIX (Amazon VPC), SUSE Linux (Amazon VPC),
    Windows (Amazon VPC)
    
-   `spot_tags` (map[string]string) - Requires spot_price to be
    set. This tells Packer to apply tags to the spot request that is issued.
    
-   `subnet_filter` (SubnetFilterOptions) - Filters used to populate the `subnet_id` field.
    Example:
    
      ``` json
      {
        "subnet_filter": {
          "filters": {
            "tag:Class": "build"
          },
          "most_free": true,
          "random": false
        }
      }
      ```
    
      This selects the Subnet with tag `Class` with the value `build`, which has
      the most free IP addresses. NOTE: This will fail unless *exactly* one
      Subnet is returned. By using `most_free` or `random` one will be selected
      from those matching the filter.
    
      -   `filters` (map of strings) - filters used to select a `subnet_id`.
          NOTE: This will fail unless *exactly* one Subnet is returned. Any
          filter described in the docs for
          [DescribeSubnets](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSubnets.html)
          is valid.
    
      -   `most_free` (boolean) - The Subnet with the most free IPv4 addresses
          will be used if multiple Subnets matches the filter.
    
      -   `random` (boolean) - A random Subnet will be used if multiple Subnets
          matches the filter. `most_free` have precendence over this.
    
      `subnet_id` take precedence over this.
    
-   `subnet_id` (string) - If using VPC, the ID of the subnet, such as
    subnet-12345def, where Packer will launch the EC2 instance. This field is
    required if you are using an non-default VPC.
    
-   `temporary_key_pair_name` (string) - The name of the temporary key pair to
    generate. By default, Packer generates a name that looks like
    `packer_<UUID>`, where &lt;UUID&gt; is a 36 character unique identifier.
    
-   `temporary_security_group_source_cidrs` ([]string) - A list of IPv4 CIDR blocks to be authorized access to the instance, when
    packer is creating a temporary security group.
    
    The default is [`0.0.0.0/0`] (i.e., allow any IPv4 source). This is only
    used when `security_group_id` or `security_group_ids` is not specified.
    
-   `user_data` (string) - User data to apply when launching the instance. Note
    that you need to be careful about escaping characters due to the templates
    being JSON. It is often more convenient to use user_data_file, instead.
    Packer will not automatically wait for a user script to finish before
    shutting down the instance this must be handled in a provisioner.
    
-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the instance.
    
-   `vpc_filter` (VpcFilterOptions) - Filters used to populate the `vpc_id` field.
    Example:
    
    ``` json
    {
      "vpc_filter": {
        "filters": {
          "tag:Class": "build",
          "isDefault": "false",
          "cidr": "/24"
        }
      }
    }
    ```
    
    This selects the VPC with tag `Class` with the value `build`, which is not
    the default VPC, and have a IPv4 CIDR block of `/24`. NOTE: This will fail
    unless *exactly* one VPC is returned.
    
    -   `filters` (map of strings) - filters used to select a `vpc_id`. NOTE:
        This will fail unless *exactly* one VPC is returned. Any filter
        described in the docs for
        [DescribeVpcs](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcs.html)
        is valid.
    
    `vpc_id` take precedence over this.
    
-   `vpc_id` (string) - If launching into a VPC subnet, Packer needs the VPC ID
    in order to create a temporary security group within the VPC. Requires
    subnet_id to be set. If this field is left blank, Packer will try to get
    the VPC ID from the subnet_id.
    
-   `windows_password_timeout` (duration string | ex: "1h5m2s") - The timeout for waiting for a Windows
    password for Windows instances. Defaults to 20 minutes. Example value:
    10m
    
-   `ssh_interface` (string) - SSH Interface