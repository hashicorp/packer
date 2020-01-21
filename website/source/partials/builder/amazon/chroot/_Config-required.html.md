<!-- Code generated from the comments of the Config struct in builder/amazon/chroot/builder.go; DO NOT EDIT MANUALLY -->

-   `source_ami` (string) - The source AMI whose root volume will be copied and provisioned on the
    currently running instance. This must be an EBS-backed AMI with a root
    volume snapshot that you have access to. Note: this is not used when
    from_scratch is set to true.
    