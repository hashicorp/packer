---
description: |
    Communicators are the mechanism Packer uses to upload files, execute scripts,
    etc. with the machine being created.
layout: docs
page_title: 'Communicators - Templates'
sidebar_current: 'docs-templates-communicators'
---

# Template Communicators

Communicators are the mechanism Packer uses to upload files, execute scripts,
etc. with the machine being created.

Communicators are configured within the
[builder](/docs/templates/builders.html) section. Packer currently supports
three kinds of communicators:

-   `none` - No communicator will be used. If this is set, most provisioners
    also can't be used.

-   `ssh` - An SSH connection will be established to the machine. This is
    usually the default.

-   `winrm` - A WinRM connection will be established.

In addition to the above, some builders have custom communicators they can use.
For example, the Docker builder has a "docker" communicator that uses
`docker exec` and `docker cp` to execute scripts and copy files.

## Using a Communicator

By default, the SSH communicator is usually used. Additional configuration may
not even be necessary, since some builders such as Amazon automatically
configure everything.

However, to specify a communicator, you set the `communicator` key within a
build. Multiple builds can have different communicators. Example:

``` json
{
  "builders": [
    {
      "type": "amazon-ebs",
      "communicator": "ssh"
    }
  ]
}
```

After specifying the `communicator`, you can specify a number of other
configuration parameters for that communicator. These are documented below.

## SSH Communicator

The SSH communicator connects to the host via SSH. If you have an SSH agent
configured on the host running Packer, and SSH agent authentication is enabled
in the communicator config, Packer will automatically forward the SSH agent to
the remote host.

The SSH communicator has the following options:

-   `ssh_agent_auth` (boolean) - If `true`, the local SSH agent will be used to
    authenticate connections to the remote host. Defaults to `false`.

-   `ssh_bastion_agent_auth` (boolean) - If `true`, the local SSH agent will be
    used to authenticate with the bastion host. Defaults to `false`.

-   `ssh_bastion_host` (string) - A bastion host to use for the actual SSH
    connection.

-   `ssh_bastion_password` (string) - The password to use to authenticate with
    the bastion host.

-   `ssh_bastion_port` (number) - The port of the bastion host. Defaults to
    `22`.

-   `ssh_bastion_private_key_file` (string) - Path to a PEM encoded private
    key file to use to authenticate with the bastion host. The `~` can be used
    in path and will be expanded to the home directory of current user.

-   `ssh_bastion_username` (string) - The username to connect to the bastion
    host.

-   `ssh_clear_authorized_keys` (boolean) - If true, Packer will attempt to
    remove its temporary key from `~/.ssh/authorized_keys` and
    `/root/.ssh/authorized_keys`. This is a mostly cosmetic option, since
    Packer will delete the temporary private key from the host system
    regardless of whether this is set to true (unless the user has set the
    `-debug` flag). Defaults to "false"; currently only works on guests with
    `sed` installed.

-   `ssh_disable_agent_forwarding` (boolean) - If true, SSH agent forwarding
    will be disabled. Defaults to `false`.

-   `ssh_file_transfer_method` (`scp` or `sftp`) - How to transfer files,
    Secure copy (default) or SSH File Transfer Protocol.

-   `ssh_handshake_attempts` (number) - The number of handshakes to attempt
    with SSH once it can connect. This defaults to `10`.

-   `ssh_host` (string) - The address to SSH to. This usually is automatically
    configured by the builder.

-   `ssh_keep_alive_interval` (string) - How often to send "keep alive"
    messages to the server. Set to a negative value (`-1s`) to disable. Example
    value: `10s`. Defaults to `5s`.

-   `ssh_password` (string) - A plaintext password to use to authenticate with
    SSH.

-   `ssh_port` (number) - The port to connect to SSH. This defaults to `22`.

-   `ssh_private_key_file` (string) - Path to a PEM encoded private key file to
    use to authenticate with SSH. The `~` can be used in path and will be
    expanded to the home directory of current user.

-   `ssh_proxy_host` (string) - A SOCKS proxy host to use for SSH connection

-   `ssh_proxy_password` (string) - The password to use to authenticate with
    the proxy server. Optional.

-   `ssh_proxy_port` (number) - A port of the SOCKS proxy. Defaults to `1080`.

-   `ssh_proxy_username` (string) - The username to authenticate with the proxy
    server. Optional.

-   `ssh_pty` (boolean) - If `true`, a PTY will be requested for the SSH
    connection. This defaults to `false`.

-   `ssh_read_write_timeout` (string) - The amount of time to wait for a remote
    command to end. This might be useful if, for example, packer hangs on a
    connection after a reboot. Example: `5m`. Disabled by default.

-   `ssh_timeout` (string) - The time to wait for SSH to become available.
    Packer uses this to determine when the machine has booted so this is
    usually quite long. Example value: `10m`.

-   `ssh_username` (string) - The username to connect to SSH with. Required if
    using SSH.

### SSH Communicator Details

Packer will only use one authentication method, either `publickey` or if
`ssh_password` is used packer will offer `password` and `keyboard-interactive`
both sending the password. In other words Packer will not work with *sshd*
configured with more than one configured authentication method using
`AuthenticationMethods`.

Packer supports the following ciphers:

-   aes128-ctr
-   aes192-ctr
-   aes256-ctr
-   arcfour128
-   arcfour256
-   arcfour
-   <es128-gcm@openssh.com>
-   <acha20-poly1305@openssh.com>

And the following MACs:

-   hmac-sha1
-   hmac-sha1-96
-   hmac-sha2-256
-   <hmac-sha2-256-etm@openssh.com>

## WinRM Communicator

The WinRM communicator has the following options.

-   `winrm_host` (string) - The address for WinRM to connect to.
    
    NOTE: If using an Amazon EBS builder, you can specify the interface WinRM connects to via [`ssh_interface`](https://www.packer.io/docs/builders/amazon-ebs.html#ssh_interface)

-   `winrm_insecure` (boolean) - If `true`, do not check server certificate
    chain and host name.

-   `winrm_password` (string) - The password to use to connect to WinRM.

-   `winrm_port` (number) - The WinRM port to connect to. This defaults to
    `5985` for plain unencrypted connection and `5986` for SSL when
    `winrm_use_ssl` is set to true.

-   `winrm_timeout` (string) - The amount of time to wait for WinRM to become
    available. This defaults to `30m` since setting up a Windows machine
    generally takes a long time.

-   `winrm_use_ntlm` (boolean) - If `true`, NTLMv2 authentication (with session
    security) will be used for WinRM, rather than default (basic
    authentication), removing the requirement for basic authentication to be
    enabled within the target guest. Further reading for remote connection
    authentication can be found
    [here](https://msdn.microsoft.com/en-us/library/aa384295(v=vs.85).aspx).

-   `winrm_use_ssl` (boolean) - If `true`, use HTTPS for WinRM.

-   `winrm_username` (string) - The username to use to connect to WinRM.
