<!-- Code generated from the comments of the ShutdownConfig struct in builder/virtualbox/common/shutdown_config.go; DO NOT EDIT MANUALLY -->

-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine unless a
    shutdown command takes place inside script so this may safely be omitted. If
    one or more scripts require a reboot it is suggested to leave this blank
    since reboots may fail and specify the final shutdown command in your
    last script.
    
-   `shutdown_timeout` (string) - The amount of time to wait after executing the
    shutdown_command for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    5m or five minutes.
    
-   `post_shutdown_delay` (string) - The amount of time to wait after shutting
    down the virtual machine. If you get the error
    Error removing floppy controller, you might need to set this to 5m
    or so. By default, the delay is 0s or disabled.
    