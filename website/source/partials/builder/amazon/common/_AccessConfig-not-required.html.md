<!-- Code generated from the comments of the AccessConfig struct in builder/amazon/common/access_config.go; DO NOT EDIT MANUALLY -->

-   `custom_endpoint_ec2` (string) - This option is useful if you use a cloud
    provider whose API is compatible with aws EC2. Specify another endpoint
    like this https://ec2.custom.endpoint.com.
    
-   `decode_authorization_messages` (bool) - Enable automatic decoding of any encoded authorization (error) messages
    using the `sts:DecodeAuthorizationMessage` API. Note: requires that the
    effective user/role have permissions to `sts:DecodeAuthorizationMessage`
    on resource `*`. Default `false`.
    
-   `insecure_skip_tls_verify` (bool) - This allows skipping TLS
    verification of the AWS EC2 endpoint. The default is false.
    
-   `max_retries` (int) - This is the maximum number of times an API call is retried, in the case
    where requests are being throttled or experiencing transient failures.
    The delay between the subsequent API calls increases exponentially.
    
-   `mfa_code` (string) - The MFA
    [TOTP](https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm)
    code. This should probably be a user variable since it changes all the
    time.
    
-   `profile` (string) - The profile to use in the shared credentials file for
    AWS. See Amazon's documentation on [specifying
    profiles](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-profiles)
    for more details.
    
-   `skip_region_validation` (bool) - Set to true if you want to skip
    validation of the ami_regions configuration option. Default false.
    
-   `skip_metadata_api_check` (bool) - Skip Metadata Api Check
-   `token` (string) - The access token to use. This is different from the
    access key and secret key. If you're not sure what this is, then you
    probably don't need it. This will also be read from the AWS_SESSION_TOKEN
    environmental variable.
    
-   `vault_aws_engine` (VaultAWSEngineOptions) - Get credentials from Hashicorp Vault's aws secrets engine. You must
    already have created a role to use. For more information about
    generating credentials via the Vault engine, see the [Vault
    docs.](https://www.vaultproject.io/api/secret/aws/index.html#generate-credentials)
    If you set this flag, you must also set the below options:
    -   `name` (string) - Required. Specifies the name of the role to generate
        credentials against. This is part of the request URL.
    -   `engine_name` (string) - The name of the aws secrets engine. In the
        Vault docs, this is normally referred to as "aws", and Packer will
        default to "aws" if `engine_name` is not set.
    -   `role_arn` (string)- The ARN of the role to assume if credential\_type
        on the Vault role is assumed\_role. Must match one of the allowed role
        ARNs in the Vault role. Optional if the Vault role only allows a single
        AWS role ARN; required otherwise.
    -   `ttl` (string) - Specifies the TTL for the use of the STS token. This
        is specified as a string with a duration suffix. Valid only when
        credential\_type is assumed\_role or federation\_token. When not
        specified, the default\_sts\_ttl set for the role will be used. If that
        is also not set, then the default value of 3600s will be used. AWS
        places limits on the maximum TTL allowed. See the AWS documentation on
        the DurationSeconds parameter for AssumeRole (for assumed\_role
        credential types) and GetFederationToken (for federation\_token
        credential types) for more details.
    
    ``` json
    {
        "vault_aws_engine": {
            "name": "myrole",
            "role_arn": "myarn",
            "ttl": "3600s"
        }
    }
    ```
    