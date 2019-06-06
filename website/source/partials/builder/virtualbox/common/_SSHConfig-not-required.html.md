<!-- Code generated from the comments of the SSHConfig struct in builder/virtualbox/common/ssh_config.go; DO NOT EDIT MANUALLY -->

-   `ssh_host_port_min` (int) - The minimum and
    maximum port to use for the SSH port on the host machine which is forwarded
    to the SSH port on the guest machine. Because Packer often runs in parallel,
    Packer will choose a randomly available port in this range to use as the
    host port. By default this is 2222 to 4444.
    
-   `ssh_host_port_max` (int) - SSH Host Port Max
-   `ssh_skip_nat_mapping` (bool) - Defaults to false. When enabled, Packer
    does not setup forwarded port mapping for SSH requests and uses ssh_port
    on the host to communicate to the virtual machine.
    
-   `ssh_wait_timeout` (time.Duration) - These are deprecated, but we keep them around for BC
    TODO(@mitchellh): remove
    