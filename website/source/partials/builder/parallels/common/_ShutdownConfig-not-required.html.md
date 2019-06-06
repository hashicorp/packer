<!-- Code generated from the comments of the ShutdownConfig struct in builder/parallels/common/shutdown_config.go; DO NOT EDIT MANUALLY -->

-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine.
    
-   `shutdown_timeout` (string) - The amount of time to wait after executing the
    shutdown_command for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    "5m", or five minutes.
    