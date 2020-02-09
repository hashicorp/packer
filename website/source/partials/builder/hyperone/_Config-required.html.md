<!-- Code generated from the comments of the Config struct in builder/hyperone/config.go; DO NOT EDIT MANUALLY -->

-   `token` (string) - The authentication token used to access your account.
    This can be either a session token or a service account token.
    If not defined, the builder will attempt to find it in the following order:
    
-   `project` (string) - The id or name of the project. This field is required
    only if using session tokens. It should be skipped when using service
    account authentication.
    
-   `source_image` (string) - ID or name of the image to launch server from.
    
-   `vm_type` (string) - ID or name of the type this server should be created with.
    
-   `disk_size` (float32) - Size of the created disk, in GiB.
    