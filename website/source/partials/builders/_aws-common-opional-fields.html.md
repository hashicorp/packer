-   `kms_key_id` (string) - ID, alias or ARN of the KMS key to use for boot
    volume encryption. This only applies to the main `region`, other regions
    where the AMI will be copied will be encrypted by the default EBS KMS key.
    For valid formats see *KmsKeyId* in the [AWS API docs -
    CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html).
    This field is validated by Packer, when using an alias, you will have to
    prefix `kms_key_id` with `alias/`.