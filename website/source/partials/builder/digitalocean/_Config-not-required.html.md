<!-- Code generated from the comments of the Config struct in builder/digitalocean/config.go; DO NOT EDIT MANUALLY -->

-   `api_url` (string) - Non standard api endpoint URL. Set this if you are
    using a DigitalOcean API compatible service. It can also be specified via
    environment variable DIGITALOCEAN_API_URL.
    
-   `private_networking` (bool) - Set to true to enable private networking
    for the droplet being created. This defaults to false, or not enabled.
    
-   `monitoring` (bool) - Set to true to enable monitoring for the droplet
    being created. This defaults to false, or not enabled.
    
-   `ipv6` (bool) - Set to true to enable ipv6 for the droplet being
    created. This defaults to false, or not enabled.
    
-   `snapshot_name` (string) - The name of the resulting snapshot that will
    appear in your account. Defaults to "packer-{{timestamp}}" (see
    configuration templates for more info).
    
-   `snapshot_regions` ([]string) - The regions of the resulting
    snapshot that will appear in your account.
    
-   `state_timeout` (time.Duration) - The time to wait, as a duration string, for a
    droplet to enter a desired state (such as "active") before timing out. The
    default state timeout is "6m".
    
-   `droplet_name` (string) - The name assigned to the droplet. DigitalOcean
    sets the hostname of the machine to this value.
    
-   `user_data` (string) - User data to launch with the Droplet. Packer will
    not automatically wait for a user script to finish before shutting down the
    instance this must be handled in a provisioner.
    
-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the Droplet.
    
-   `tags` ([]string) - Tags to apply to the droplet when it is created
    