<!-- Code generated from the comments of the OutputConfig struct in builder/hyperv/common/output_config.go; DO NOT EDIT MANUALLY -->

-   `output_directory` (string) - This setting specifies the directory that
    artifacts from the build, such as the virtual machine files and disks,
    will be output to. The path to the directory may be relative or
    absolute. If relative, the path is relative to the working directory
    packer is executed from. This directory must not exist or, if
    created, must be empty prior to running the builder. By default this is
    "output-BUILDNAME" where "BUILDNAME" is the name of the build.
    