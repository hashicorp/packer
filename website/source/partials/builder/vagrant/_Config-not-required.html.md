<!-- Code generated from the comments of the Config struct in builder/vagrant/builder.go; DO NOT EDIT MANUALLY -->

-   `output_dir` (string) - The directory to create that will contain your output box. We always
    create this directory and run from inside of it to prevent Vagrant init
    collisions. If unset, it will be set to packer- plus your buildname.
    
-   `checksum` (string) - The checksum for the .box file. The type of the checksum is specified
    with checksum_type, documented below.
    
-   `checksum_type` (string) - The type of the checksum specified in checksum. Valid values are none,
    md5, sha1, sha256, or sha512. Although the checksum will not be verified
    when checksum_type is set to "none", this is not recommended since OVA
    files can be very large and corruption does happen from time to time.
    
-   `box_name` (string) - if your source_box is a boxfile that we need to add to Vagrant, this is
    the name to give it. If left blank, will default to "packer_" plus your
    buildname.
    
-   `insert_key` (bool) - If true, Vagrant will automatically insert a keypair to use for SSH,
    replacing Vagrant's default insecure key inside the machine if detected.
    By default, Packer sets this to false.
    
-   `provider` (string) - The vagrant provider.
    This parameter is required when source_path have more than one provider,
    or when using vagrant-cloud post-processor. Defaults to unset.
    
-   `communicator` (string) - Communicator
-   `vagrantfile_template` (string) - What vagrantfile to use
    
-   `teardown_method` (string) - Whether to halt, suspend, or destroy the box when the build has
    completed. Defaults to "halt"
    
-   `box_version` (string) - What box version to use when initializing Vagrant.
    
-   `template` (string) - a path to a golang template for a vagrantfile. Our default template can
    be found here. The template variables available to you are
    {{ .BoxName }}, {{ .SyncedFolder }}, and {{.InsertKey}}, which
    correspond to the Packer options box_name, synced_folder, and insert_key.
    
-   `synced_folder` (string) - Synced Folder
-   `skip_add` (bool) - Don't call "vagrant add" to add the box to your local environment; this
    is necessary if you want to launch a box that is already added to your
    vagrant environment.
    
-   `add_cacert` (string) - Equivalent to setting the
    --cacert
    option in vagrant add; defaults to unset.
    
-   `add_capath` (string) - Equivalent to setting the
    --capath option
    in vagrant add; defaults to unset.
    
-   `add_cert` (string) - Equivalent to setting the
    --cert option in
    vagrant add; defaults to unset.
    
-   `add_clean` (bool) - Equivalent to setting the
    --clean flag in
    vagrant add; defaults to unset.
    
-   `add_force` (bool) - Equivalent to setting the
    --force flag in
    vagrant add; defaults to unset.
    
-   `add_insecure` (bool) - Equivalent to setting the
    --insecure flag in
    vagrant add; defaults to unset.
    
-   `skip_package` (bool) - if true, Packer will not call vagrant package to
    package your base box into its own standalone .box file.
    
-   `output_vagrantfile` (string) - Output Vagrantfile
-   `package_include` ([]string) - Package Include