<!-- Code generated from the comments of the Config struct in builder/azure/common/client/config.go; DO NOT EDIT MANUALLY -->

-   `cloud_environment_name` (string) - One of Public, China, Germany, or
    USGovernment. Defaults to Public. Long forms such as
    USGovernmentCloud and AzureUSGovernmentCloud are also supported.
    
-   `client_id` (string) - The application ID of the AAD Service Principal.
    Requires either `client_secret`, `client_cert_path` or `client_jwt` to be set as well.
    
-   `client_secret` (string) - A password/secret registered for the AAD SP.
    
-   `client_cert_path` (string) - The path to a pem-encoded certificate that will be used to authenticate
    as the specified AAD SP.
    
-   `client_jwt` (string) - A JWT bearer token for client auth (RFC 7523, Sec. 2.2) that will be used
    to authenticate the AAD SP. Provides more control over token the expiration
    when using certificate authentication than when using `client_cert_path`.
    
-   `object_id` (string) - The object ID for the AAD SP. Optional, will be derived from the oAuth token if left empty.
    
-   `tenant_id` (string) - The Active Directory tenant identifier with which your `client_id` and
    `subscription_id` are associated. If not specified, `tenant_id` will be
    looked up using `subscription_id`.
    
-   `subscription_id` (string) - The subscription to use.
    