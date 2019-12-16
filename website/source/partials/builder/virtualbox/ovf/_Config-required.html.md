<!-- Code generated from the comments of the Config struct in builder/virtualbox/ovf/config.go; DO NOT EDIT MANUALLY -->

-   `checksum` (string) - The checksum for the source_path file. The
    algorithm to use when computing the checksum can be optionally specified
    with checksum_type. When checksum_type is not set packer will guess the
    checksumming type based on checksum length. checksum can be also be a
    file or an URL, in which case checksum_type must be set to file; the
    go-getter will download it and use the first hash found.
    
-   `source_path` (string) - The filepath or URL to an OVF or OVA file that acts as the
    source of this build.
    