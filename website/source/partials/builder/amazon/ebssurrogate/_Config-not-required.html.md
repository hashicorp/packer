<!-- Code generated from the comments of the Config struct in builder/amazon/ebssurrogate/builder.go; DO NOT EDIT MANUALLY -->

-   `run_volume_tags` (awscommon.TagMap) - Tags to apply to the volumes that are *launched* to create the AMI.
    These tags are *not* applied to the resulting AMI unless they're
    duplicated in `tags`. This is a [template
    engine](/docs/templates/engine.html), see [Build template
    data](#build-template-data) for more information.
    
-   `ami_architecture` (string) - what architecture to use when registering the
    final AMI; valid options are "x86_64" or "arm64". Defaults to "x86_64".
    