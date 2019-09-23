<!-- Code generated from the comments of the Config struct in builder/cloudstack/config.go; DO NOT EDIT MANUALLY -->

-   `api_url` (string) - The CloudStack API endpoint we will connect to. It can
    also be specified via environment variable CLOUDSTACK_API_URL, if set.
    
-   `api_key` (string) - The API key used to sign all API requests. It can also
    be specified via environment variable CLOUDSTACK_API_KEY, if set.
    
-   `secret_key` (string) - The secret key used to sign all API requests. It
    can also be specified via environment variable CLOUDSTACK_SECRET_KEY, if
    set.
    
-   `network` (string) - The name or ID of the network to connect the instance
    to.
    
-   `service_offering` (string) - The name or ID of the service offering used
    for the instance.
    
-   `source_iso` (string) - The name or ID of an ISO that will be mounted
    before booting the instance. This option is mutually exclusive with
    source_template. When using source_iso, both disk_offering and
    hypervisor are required.
    
-   `source_template` (string) - The name or ID of the template used as base
    template for the instance. This option is mutually exclusive with
    source_iso.
    
-   `zone` (string) - The name or ID of the zone where the instance will be
    created.
    
-   `template_os` (string) - The name or ID of the template OS for the new
    template that will be created.
    