<!-- Code generated from the comments of the SSH struct in helper/communicator/config.go; DO NOT EDIT MANUALLY -->

-   `ssh_host` (string) - The address to SSH to. This usually is automatically configured by the
    builder.
    
-   `ssh_port` (int) - The port to connect to SSH. This defaults to `22`.
    
-   `ssh_username` (string) - The username to connect to SSH with. Required if using SSH.
    
-   `ssh_password` (string) - A plaintext password to use to authenticate with SSH.
    
-   `ssh_keypair_name` (string) - If specified, this is the key that will be used for SSH with the
    machine. The key must match a key pair name loaded up into Amazon EC2.
    By default, this is blank, and Packer will generate a temporary keypair
    unless [`ssh_password`](../templates/communicator.html#ssh_password) is
    used.
    [`ssh_private_key_file`](../templates/communicator.html#ssh_private_key_file)
    or `ssh_agent_auth` must be specified when `ssh_keypair_name` is
    utilized.
    
-   `temporary_key_pair_name` (string) - SSH Temporary Key Pair Name
-   `ssh_clear_authorized_keys` (bool) - If true, Packer will attempt to remove its temporary key from
    `~/.ssh/authorized_keys` and `/root/.ssh/authorized_keys`. This is a
    mostly cosmetic option, since Packer will delete the temporary private
    key from the host system regardless of whether this is set to true
    (unless the user has set the `-debug` flag). Defaults to "false";
    currently only works on guests with `sed` installed.
    
-   `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
    The `~` can be used in path and will be expanded to the home directory
    of current user.
    
-   `ssh_interface` (string) - One of `public_ip`, `private_ip`, `public_dns`, or `private_dns`. If
    set, either the public IP address, private IP address, public DNS name
    or private DNS name will used as the host for SSH. The default behaviour
    if inside a VPC is to use the public IP address if available, otherwise
    the private IP address will be used. If not in a VPC the public DNS name
    will be used. Also works for WinRM.
    
    Where Packer is configured for an outbound proxy but WinRM traffic
    should be direct, `ssh_interface` must be set to `private_dns` and
    `<region>.compute.internal` included in the `NO_PROXY` environment
    variable.
    
-   `ssh_ip_version` (string) - SSHIP Version
-   `ssh_pty` (bool) - If `true`, a PTY will be requested for the SSH connection. This defaults
    to `false`.
    
-   `ssh_timeout` (time.Duration) - The time to wait for SSH to become available. Packer uses this to
    determine when the machine has booted so this is usually quite long.
    Example value: `10m`.
    
-   `ssh_agent_auth` (bool) - If true, the local SSH agent will be used to authenticate connections to
    the source instance. No temporary keypair will be created, and the
    values of `ssh_password` and `ssh_private_key_file` will be ignored. To
    use this option with a key pair already configured in the source AMI,
    leave the `ssh_keypair_name` blank. To associate an existing key pair in
    AWS with the source instance, set the `ssh_keypair_name` field to the
    name of the key pair.
    
-   `ssh_disable_agent_forwarding` (bool) - If true, SSH agent forwarding will be disabled. Defaults to `false`.
    
-   `ssh_handshake_attempts` (int) - The number of handshakes to attempt with SSH once it can connect. This
    defaults to `10`.
    
-   `ssh_bastion_host` (string) - A bastion host to use for the actual SSH connection.
    
-   `ssh_bastion_port` (int) - The port of the bastion host. Defaults to `22`.
    
-   `ssh_bastion_agent_auth` (bool) - If `true`, the local SSH agent will be used to authenticate with the
    bastion host. Defaults to `false`.
    
-   `ssh_bastion_username` (string) - The username to connect to the bastion host.
    
-   `ssh_bastion_password` (string) - The password to use to authenticate with the bastion host.
    
-   `ssh_bastion_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with the
    bastion host. The `~` can be used in path and will be expanded to the
    home directory of current user.
    
-   `ssh_file_transfer_method` (string) - `scp` or `sftp` - How to transfer files, Secure copy (default) or SSH
    File Transfer Protocol.
    
-   `ssh_proxy_host` (string) - A SOCKS proxy host to use for SSH connection
    
-   `ssh_proxy_port` (int) - A port of the SOCKS proxy. Defaults to `1080`.
    
-   `ssh_proxy_username` (string) - The optional username to authenticate with the proxy server.
    
-   `ssh_proxy_password` (string) - The optional password to use to authenticate with the proxy server.
    
-   `ssh_keep_alive_interval` (time.Duration) - How often to send "keep alive" messages to the server. Set to a negative
    value (`-1s`) to disable. Example value: `10s`. Defaults to `5s`.
    
-   `ssh_read_write_timeout` (time.Duration) - The amount of time to wait for a remote command to end. This might be
    useful if, for example, packer hangs on a connection after a reboot.
    Example: `5m`. Disabled by default.
    