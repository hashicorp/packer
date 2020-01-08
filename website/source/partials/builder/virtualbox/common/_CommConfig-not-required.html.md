<!-- Code generated from the comments of the CommConfig struct in builder/virtualbox/common/comm_config.go; DO NOT EDIT MANUALLY -->

-   `host_port_min` (int) - The minimum port to use for the Communicator port on the host machine which is forwarded
    to the SSH or WinRM port on the guest machine. By default this is 2222.
    
-   `host_port_max` (int) - The maximum port to use for the Communicator port on the host machine which is forwarded
    to the SSH or WinRM port on the guest machine. Because Packer often runs in parallel,
    Packer will choose a randomly available port in this range to use as the
    host port. By default this is 4444.
    
-   `skip_nat_mapping` (bool) - Defaults to false. When enabled, Packer
    does not setup forwarded port mapping for communicator (SSH or WinRM) requests and uses ssh_port or winrm_port
    on the host to communicate to the virtual machine.
    