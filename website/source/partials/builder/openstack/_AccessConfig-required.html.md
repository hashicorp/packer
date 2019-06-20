<!-- Code generated from the comments of the AccessConfig struct in builder/openstack/access_config.go; DO NOT EDIT MANUALLY -->

-   `username` (string) - The username or id used to connect to the OpenStack service. If not
    specified, Packer will use the environment variable OS_USERNAME or
    OS_USERID, if set. This is not required if using access token or
    application credential instead of password, or if using cloud.yaml.
    
-   `password` (string) - The password used to connect to the OpenStack service. If not specified,
    Packer will use the environment variables OS_PASSWORD, if set. This is
    not required if using access token or application credential instead of
    password, or if using cloud.yaml.
    
-   `identity_endpoint` (string) - The URL to the OpenStack Identity service. If not specified, Packer will
    use the environment variables OS_AUTH_URL, if set. This is not required
    if using cloud.yaml.
    