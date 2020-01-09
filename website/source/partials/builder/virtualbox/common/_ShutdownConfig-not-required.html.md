<!-- Code generated from the comments of the ShutdownConfig struct in builder/virtualbox/common/shutdown_config.go; DO NOT EDIT MANUALLY -->

-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine unless a
    shutdown command takes place inside script so this may safely be omitted. If
    one or more scripts require a reboot it is suggested to leave this blank
    since reboots may fail and specify the final shutdown command in your
    last script.
    
-   `shutdown_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait after executing the
    shutdown_command for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    5m or five minutes.
    
-   `post_shutdown_delay` (duration string | ex: "1h5m2s") - The amount of time to wait after shutting
    down the virtual machine. If you get the error
    Error removing floppy controller, you might need to set this to 5m
    or so. By default, the delay is 0s or disabled.
    
-   `disable_shutdown` (bool) - Packer normally halts the virtual machine after all provisioners have
    run when no `shutdown_command` is defined.  If this is set to `true`, Packer
    *will not* halt the virtual machine but will assume that you will send the stop
    signal yourself through the preseed.cfg or your final provisioner.
    Packer will wait for a default of 5 minutes until the virtual machine is shutdown.
    The timeout can be changed using `shutdown_timeout` option.
    
-   `acpi_shutdown` (bool) - If it's set to true, it will shutdown the VM via power button. It could be a good option
    when keeping the machine state is necessary after shutting it down.
    