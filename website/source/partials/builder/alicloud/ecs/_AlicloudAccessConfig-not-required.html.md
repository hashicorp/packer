<!-- Code generated from the comments of the AlicloudAccessConfig struct in builder/alicloud/ecs/access_config.go; DO NOT EDIT MANUALLY -->

-   `access_key` (string) - This is the Alicloud access key. It must be provided when profile not exist, but it can also be
    sourced from the ALICLOUD_ACCESS_KEY environment variable.
    
-   `secret_key` (string) - This is the Alicloud secret key. It must be provided when profile not exist, but it can also be
    sourced from the ALICLOUD_SECRET_KEY environment variable.
    
-   `region` (string) - This is the Alicloud region. It must be provided when profile not exist, but it can also be
    sourced from the ALICLOUD_REGION environment variables.
    
-   `skip_region_validation` (bool) - The region validation can be skipped if this value is true, the default
    value is false.
    
-   `skip_image_validation` (bool) - The image validation can be skipped if this value is true, the default
    value is false.
    
-   `profile` (string) - This is th Alicloud profile. If access_key not exist, is must be provided, but it can also be
    sourced from the ALICLOUD_PROFILE environment variables.
    
-   `shared_credentials_file` (string) - This is the Alicloud shared credentials file path. If this file path exist, os will read access key
    and secret key from this file.
    
-   `security_token` (string) - STS access token, can be set through template or by exporting as
    environment variable such as `export SECURITY_TOKEN=value`.
    