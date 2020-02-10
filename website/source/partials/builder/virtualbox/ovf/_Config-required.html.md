<!-- Code generated from the comments of the Config struct in builder/virtualbox/ovf/config.go; DO NOT EDIT MANUALLY -->

-   `checksum` (string) - The checksum for the source_path file. The type of the checksum is
    specified within the checksum as a prefix, ex: "md5:{$checksum}". The
    type of the checksum can also be omitted and Packer will try to infer it
    based on string length. Valid values are "none", "{$checksum}",
    "{$path}" "md5:{$checksum}", "sha1:{$checksum}", "sha256:{$checksum}",
    "sha512:{$checksum}" or "file:{$checksum}". Although the checksum will
    not be verified when it is set to "none", this is not recommended since
    these files can be very large and corruption does happen from time to
    time.
    
-   `source_path` (string) - The filepath or URL to an OVF or OVA file that acts as the
    source of this build.
    