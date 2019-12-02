<!-- Code generated from the comments of the ISOConfig struct in common/iso_config.go; DO NOT EDIT MANUALLY -->

-   `iso_checksum` (string) - The checksum for the ISO file or virtual hard drive file. The algorithm
    to use when computing the checksum will be determined automatically
    based on `iso_checksum` length. `iso_checksum` can be also be a file or
    an URL, in which case iso_checksum must be prefixed with `file:`; the
    go-getter will download it and use the first hash found.
    
    `iso_checksum` can be set to `"none"` if you want no checksumming
    operation to be run.
    
-   `iso_url` (string) - A URL to the ISO containing the installation image or virtual hard drive
    (VHD or VHDX) file to clone.
    