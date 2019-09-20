<!-- Code generated from the comments of the ClientConfig struct in builder/azure/arm/clientconfig.go; DO NOT EDIT MANUALLY -->

-   `cloud_environment_name` (string) - One of Public, China, Germany, or
    USGovernment. Defaults to Public. Long forms such as
    USGovernmentCloud and AzureUSGovernmentCloud are also supported.
    
-   `client_id` (string) - Client ID
    
-   `client_secret` (string) - Client secret/password
    
-   `client_cert_path` (string) - Certificate path for client auth
    
-   `client_jwt` (string) - JWT bearer token for client auth (RFC 7523, Sec. 2.2)
    
-   `object_id` (string) - Object ID
-   `tenant_id` (string) - The account identifier with which your client_id and
    subscription_id are associated. If not specified, tenant_id will be
    looked up using subscription_id.
    
-   `subscription_id` (string) - Subscription ID