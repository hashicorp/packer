<!-- Code generated from the comments of the Config struct in builder/amazon/instance/builder.go; DO NOT EDIT MANUALLY -->

-   `bundle_destination` (string) - The directory on the running instance where
    the bundled AMI will be saved prior to uploading. By default this is
    /tmp. This directory must exist and be writable.
    
-   `bundle_prefix` (string) - The prefix for files created from bundling the
    root volume. By default this is image-{{timestamp}}. The timestamp
    variable should be used to make sure this is unique, otherwise it can
    collide with other created AMIs by Packer in your account.
    
-   `bundle_upload_command` (string) - The command to use to upload the bundled
    volume. See the "custom bundle commands" section below for more
    information.
    
-   `bundle_vol_command` (string) - The command to use to bundle the volume.
    See the "custom bundle commands" section below for more information.
    
-   `x509_upload_path` (string) - The path on the remote machine where the X509
    certificate will be uploaded. This path must already exist and be writable.
    X509 certificates are uploaded after provisioning is run, so it is
    perfectly okay to create this directory as part of the provisioning
    process. Defaults to /tmp.
    