<!-- Code generated from the comments of the RunConfig struct in builder/openstack/run_config.go; DO NOT EDIT MANUALLY -->

-   `source_image` (string) - The ID or full URL to the base image to use. This
    is the image that will be used to launch a new server and provision it.
    Unless you specify completely custom SSH settings, the source image must
    have cloud-init installed so that the keypair gets assigned properly.
    
-   `source_image_name` (string) - The name of the base image to use. This is
    an alternative way of providing source_image and only either of them can
    be specified.
    
-   `source_image_filter` (ImageFilter) - The search filters for determining the base
    image to use. This is an alternative way of providing source_image and
    only one of these methods can be used. source_image will override the
    filters.
    
-   `flavor` (string) - The ID, name, or full URL for the desired flavor for
    the server to be created.
    