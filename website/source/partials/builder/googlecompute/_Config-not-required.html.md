<!-- Code generated from the comments of the Config struct in builder/googlecompute/config.go; DO NOT EDIT MANUALLY -->

-   `account_file` (string) - The JSON file containing your account
    credentials. Not required if you run Packer on a GCE instance with a
    service account. Instructions for creating the file or using service
    accounts are above.
    
-   `accelerator_type` (string) - Full or partial URL of the guest accelerator
    type. GPU accelerators can only be used with
    "on_host_maintenance": "TERMINATE" option set. Example:
    "projects/project_id/zones/europe-west1-b/acceleratorTypes/nvidia-tesla-k80"
    
-   `accelerator_count` (int64) - Number of guest accelerator cards to add to
    the launched instance.
    
-   `address` (string) - The name of a pre-allocated static external IP
    address. Note, must be the name and not the actual IP address.
    
-   `disable_default_service_account` (bool) - If true, the default service
    account will not be used if service_account_email is not specified. Set
    this value to true and omit service_account_email to provision a VM with
    no service account.
    
-   `disk_name` (string) - The name of the disk, if unset the instance name
    will be used.
    
-   `disk_size` (int64) - The size of the disk in GB. This defaults to 10,
    which is 10GB.
    
-   `disk_type` (string) - Type of disk used to back your instance, like
    pd-ssd or pd-standard. Defaults to pd-standard.
    
-   `image_name` (string) - The unique name of the resulting image. Defaults to
    "packer-{{timestamp}}".
    
-   `image_description` (string) - The description of the resulting image.
    
-   `image_encryption_key` (*compute.CustomerEncryptionKey) - Image encryption key to apply to the created image. Possible values:
    
-   `image_family` (string) - The name of the image family to which the
    resulting image belongs. You can create disks by specifying an image family
    instead of a specific image name. The image family always returns its
    latest image that is not deprecated.
    
-   `image_labels` (map[string]string) - Key/value pair labels to
    apply to the created image.
    
-   `image_licenses` ([]string) - Licenses to apply to the created
    image.
    
-   `instance_name` (string) - A name to give the launched instance. Beware
    that this must be unique. Defaults to "packer-{{uuid}}".
    
-   `labels` (map[string]string) - Key/value pair labels to apply to
    the launched instance.
    
-   `machine_type` (string) - The machine type. Defaults to "n1-standard-1".
    
-   `metadata` (map[string]string) - Metadata applied to the launched
    instance.
    
-   `min_cpu_platform` (string) - A Minimum CPU Platform for VM Instance.
    Availability and default CPU platforms vary across zones, based on the
    hardware available in each GCP zone.
    Details
    
-   `network` (string) - The Google Compute network id or URL to use for the
    launched instance. Defaults to "default". If the value is not a URL, it
    will be interpolated to
    projects/((network_project_id))/global/networks/((network)). This value
    is not required if a subnet is specified.
    
-   `network_project_id` (string) - The project ID for the network and
    subnetwork to use for launched instance. Defaults to project_id.
    
-   `omit_external_ip` (bool) - If true, the instance will not have an
    external IP. use_internal_ip must be true if this property is true.
    
-   `on_host_maintenance` (string) - Sets Host Maintenance Option. Valid
    choices are MIGRATE and TERMINATE. Please see GCE Instance Scheduling
    Options,
    as not all machine_types support MIGRATE (i.e. machines with GPUs). If
    preemptible is true this can only be TERMINATE. If preemptible is false,
    it defaults to MIGRATE
    
-   `preemptible` (bool) - If true, launch a preemptible instance.
    
-   `state_timeout` (string) - The time to wait for instance state changes.
    Defaults to "5m".
    
-   `region` (string) - The region in which to launch the instance. Defaults to
    the region hosting the specified zone.
    
-   `scopes` ([]string) - The service account scopes for launched
    instance. Defaults to:
    
-   `service_account_email` (string) - The service account to be used for
    launched instance. Defaults to the project's default service account unless
    disable_default_service_account is true.
    
-   `source_image_project_id` (string) - The project ID of the project
    containing the source image.
    
-   `startup_script_file` (string) - The path to a startup script to run on the
    VM from which the image will be made.
    
-   `subnetwork` (string) - The Google Compute subnetwork id or URL to use for
    the launched instance. Only required if the network has been created with
    custom subnetting. Note, the region of the subnetwork must match the
    region or zone in which the VM is launched. If the value is not a URL,
    it will be interpolated to
    projects/((network_project_id))/regions/((region))/subnetworks/((subnetwork))
    
-   `tags` ([]string) - Assign network tags to apply firewall rules to
    VM instance.
    
-   `use_internal_ip` (bool) - If true, use the instance's internal IP
    instead of its external IP during building.
    