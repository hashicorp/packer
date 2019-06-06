<!-- Code generated from the comments of the VBoxManageConfig struct in builder/virtualbox/common/vboxmanage_config.go; DO NOT EDIT MANUALLY -->

-   `vboxmanage` ([][]string) - Custom VBoxManage commands to
    execute in order to further customize the virtual machine being created. The
    value of this is an array of commands to execute. The commands are executed
    in the order defined in the template. For each command, the command is
    defined itself as an array of strings, where each string represents a single
    argument on the command-line to VBoxManage (but excluding
    VBoxManage itself). Each arg is treated as a configuration
    template, where the Name
    variable is replaced with the VM name. More details on how to use
    VBoxManage are below.
    