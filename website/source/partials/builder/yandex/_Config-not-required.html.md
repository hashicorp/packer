<!-- Code generated from the comments of the Config struct in builder/yandex/config.go; DO NOT EDIT MANUALLY -->

-   `endpoint` (string) - Non standard api endpoint URL.
    
-   `service_account_key_file` (string) - Path to file with Service Account key in json format. This
    is an alternative method to authenticate to Yandex.Cloud. Alternatively you may set environment variable
    YC_SERVICE_ACCOUNT_KEY_FILE.
    
-   `service_account_id` (string) - Service account identifier to assign to instance
    
-   `disk_name` (string) - The name of the disk, if unset the instance name
    will be used.
    
-   `disk_size_gb` (int) - The size of the disk in GB. This defaults to `10`, which is 10GB.
    
-   `disk_type` (string) - Specify disk type for the launched instance. Defaults to `network-hdd`.
    
-   `image_description` (string) - The description of the resulting image.
    
-   `image_family` (string) -  The family name of the resulting image.
    
-   `image_labels` (map[string]string) - Key/value pair labels to
    apply to the created image.
    
-   `image_name` (string) - The unique name of the resulting image. Defaults to
    `packer-{{timestamp}}`.
    
-   `image_product_ids` ([]string) - License IDs that indicate which licenses are attached to resulting image.
    
-   `instance_cores` (int) - The number of cores available to the instance.
    
-   `instance_gpus` (int) - The number of GPU available to the instance.
    
-   `instance_mem_gb` (int) - The amount of memory available to the instance, specified in gigabytes.
    
-   `instance_name` (string) - The name assigned to the instance.
    
-   `labels` (map[string]string) - Key/value pair labels to apply to
    the launched instance.
    
-   `platform_id` (string) - Identifier of the hardware platform configuration for the instance. This defaults to `standard-v1`.
    
-   `max_retries` (int) - The maximum number of times an API request is being executed
    
-   `metadata` (map[string]string) - Metadata applied to the launched instance.
    
-   `metadata_from_file` (map[string]string) - Metadata applied to the launched instance. Value are file paths.
    
-   `preemptible` (bool) - Launch a preemptible instance. This defaults to `false`.
    
-   `serial_log_file` (string) - File path to save serial port output of the launched instance.
    
-   `source_image_folder_id` (string) - The ID of the folder containing the source image.
    
-   `source_image_id` (string) - The source image ID to use to create the new image
    from.
    
-   `source_image_name` (string) - The source image name to use to create the new image
    from. Name will be looked up in `source_image_folder_id`.
    
-   `subnet_id` (string) - The Yandex VPC subnet id to use for
    the launched instance. Note, the zone of the subnet must match the
    zone in which the VM is launched.
    
-   `use_ipv4_nat` (bool) - If set to true, then launched instance will have external internet
    access.
    
-   `use_ipv6` (bool) - Set to true to enable IPv6 for the instance being
    created. This defaults to `false`, or not enabled.
    
    -> **Note**: Usage of IPv6 will be available in the future.
    
-   `use_internal_ip` (bool) - If true, use the instance's internal IP address
    instead of its external IP during building.
    
-   `zone` (string) - The name of the zone to launch the instance.  This defaults to `ru-central1-a`.
    
-   `state_timeout` (duration string | ex: "1h5m2s") - The time to wait for instance state changes.
    Defaults to `5m`.
    