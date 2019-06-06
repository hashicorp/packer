<!-- Code generated from the comments of the AMIConfig struct in builder/amazon/common/ami_config.go; DO NOT EDIT MANUALLY -->

-   `ami_description` (string) - The description to set for the resulting
    AMI(s). By default this description is empty. This is a template
    engine, see Build template
    data for more information.
    
-   `ami_virtualization_type` (string) - The description to set for the resulting AMI(s). By default this
    description is empty. This is a [template
    engine](../templates/engine.html), see [Build template
    data](#build-template-data) for more information.
    
-   `ami_users` ([]string) - A list of account IDs that have access to
    launch the resulting AMI(s). By default no additional users other than the
    user creating the AMI has permissions to launch it.
    
-   `ami_groups` ([]string) - A list of groups that have access to
    launch the resulting AMI(s). By default no groups have permission to launch
    the AMI. all will make the AMI publicly accessible.
    
-   `ami_product_codes` ([]string) - A list of product codes to
    associate with the AMI. By default no product codes are associated with the
    AMI.
    
-   `ami_regions` ([]string) - A list of regions to copy the AMI to.
    Tags and attributes are copied along with the AMI. AMI copying takes time
    depending on the size of the AMI, but will generally take many minutes.
    
-   `skip_region_validation` (bool) - Set to true if you want to skip
    validation of the ami_regions configuration option. Default false.
    
-   `tags` (TagMap) - Tags applied to the AMI. This is a
    [template engine](/docs/templates/engine.html), see [Build template
    data](#build-template-data) for more information.
    
-   `ena_support` (*bool) - Enable enhanced networking (ENA but not
    SriovNetSupport) on HVM-compatible AMIs. If set, add
    ec2:ModifyInstanceAttribute to your AWS IAM policy. If false, this will
    disable enhanced networking in the final AMI as opposed to passing the
    setting through unchanged from the source. Note: you must make sure
    enhanced networking is enabled on your instance. [Amazon's
    documentation on enabling enhanced
    networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking).
    
-   `sriov_support` (bool) - Enable enhanced networking (SriovNetSupport but not ENA) on
    HVM-compatible AMIs. If true, add `ec2:ModifyInstanceAttribute` to your
    AWS IAM policy. Note: you must make sure enhanced networking is enabled
    on your instance. See [Amazon's documentation on enabling enhanced
    networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking).
    Default `false`.
    
-   `force_deregister` (bool) - Force Packer to first deregister an existing
    AMI if one with the same name already exists. Default false.
    
-   `force_delete_snapshot` (bool) - Force Packer to delete snapshots
    associated with AMIs, which have been deregistered by force_deregister.
    Default false.
    
-   `encrypt_boot` (*bool) - Whether or not to encrypt the resulting AMI when
    copying a provisioned instance to an AMI. By default, Packer will keep the
    encryption setting to what it was in the source image. Setting false will
    result in an unencrypted image, and true will result in an encrypted one.
    
-   `kms_key_id` (string) - ID, alias or ARN of the KMS key to use for boot volume encryption. This
    only applies to the main `region`, other regions where the AMI will be
    copied will be encrypted by the default EBS KMS key. For valid formats
    see *KmsKeyId* in the [AWS API docs -
    CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html).
    This field is validated by Packer, when using an alias, you will have to
    prefix `kms_key_id` with `alias/`.
    
-   `region_kms_key_ids` (map[string]string) - regions to copy the ami to, along with the custom kms key id (alias or
    arn) to use for encryption for that region. Keys must match the regions
    provided in `ami_regions`. If you just want to encrypt using a default
    ID, you can stick with `kms_key_id` and `ami_regions`. If you want a
    region to be encrypted with that region's default key ID, you can use an
    empty string `""` instead of a key id in this map. (e.g. `"us-east-1":
    ""`) However, you cannot use default key IDs if you are using this in
    conjunction with `snapshot_users` -- in that situation you must use
    custom keys. For valid formats see *KmsKeyId* in the [AWS API docs -
    CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html).
    
-   `snapshot_tags` (TagMap) - Tags to apply to snapshot.
    They will override AMI tags if already applied to snapshot. This is a
    [template engine](../templates/engine.html), see [Build template
    data](#build-template-data) for more information.
    
-   `snapshot_users` ([]string) - A list of account IDs that have
    access to create volumes from the snapshot(s). By default no additional
    users other than the user creating the AMI has permissions to create
    volumes from the backing snapshot(s).
    
-   `snapshot_groups` ([]string) - A list of groups that have access to
    create volumes from the snapshot(s). By default no groups have permission
    to create volumes from the snapshot(s). all will make the snapshot
    publicly accessible.
    