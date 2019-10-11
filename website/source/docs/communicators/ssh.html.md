---
description: |
    The SSH communicator uses SSH to upload files, execute scripts, etc. on
    the machine being created.
layout: docs
page_title: 'Communicators - SSH'
sidebar_current: 'docs-communicators-ssh'
---

# SSH Communicator

Communicators are the mechanism Packer uses to upload files, execute scripts,
etc. on the machine being created, and ar configured within the
[builder](/docs/templates/builders.html) section.

The SSH communicator does this by using the SSH protocol. It is the default
communicator for a majority of builders.

If you have an SSH agent configured on the host running Packer, and SSH agent
authentication is enabled in the communicator config, Packer will automatically
forward the SSH agent to the remote host.

## Getting Ready to Use the SSH Communicator

The SSH communicator is the default communicator for a majority of builders, but
depending on your builder it may not work "out of the box".

If you are building from a cloud image (for example, building on Amazon), there
is a good chance that your cloud provider has already preconfigured SSH on the
image for you, meaning that all you have to do is configure the communicator in
the Packer template.

However, if you are building from a brand-new and unconfigured operating system
image, you will almost always have to perform some extra work to configure SSH
on the guest machine. For most operating system distributions, this work will
be performed by a
(boot command)[/docs/builders/vmware-iso.html#boot-configuration]
that references a file which provides answers to the normally-interactive
questions you get asked when installing an operating system. The name of this
file varies by operating system; some common examples are the "preseed" file
required by Debian, the "kickstart" file required by CentOS or the
"answer file", also known as the Autounattend.xml file, required by Windows.
For simplicity's sake, we'll refer to this file as the "preseed" file in the
rest of the documentation.

If you are unfamiliar with how to use a preseed file for automatic
bootstrapping of an image, please either take a look at our [quick guides](/guides/automatic-operating-system-installs/index.html) to
image bootstrapping, or research automatic configuration for your specific
guest operating system. Knowing how to automatically initalize your operating
system is critical for being able to successfully use Packer.

## Using The SSH Communicator

To specify a communicator, you set the `communicator` key within a
build. If your template contains multiple builds, you can have a different
communicator configured for each. Here's an extremely basic example of
configuring the SSH communicator for an Amazon builder:

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

After specifying the `communicator` type, you can specify a number of other
configuration parameters for that communicator. These are documented below.

## SSH Communicator Options

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

-   `ssh_bastion_private_key_file` (string) - Path to a PEM encoded private key
    file to use to authenticate with the bastion host. The `~` can be used in
    path and will be expanded to the home directory of current user.

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

-   `ssh_local_tunnels` (array of strings) - An array of OpenSSH-style tunnels to
    create. The port is bound on the *local packer host* and connections are
    forwarded to the remote destination. Note unless `GatewayPorts=yes` is set
    in SSHD daemon, the target *must* be `localhost`. Example value:
    `3306:localhost:3306`

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

-   `ssh_remote_tunnels` (array of strings) - An array of OpenSSH-style tunnels
    to create. The port is bound on the *remote build host* and connections to it are
    forwarded to the packer host's network. Non-localhost destinations may be set here.
    Example value: `8443:git.example.com:443`

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
-   `es128-gcm@openssh.com`
-   `acha20-poly1305@openssh.com`

And the following MACs:

-   hmac-sha1
-   hmac-sha1-96
-   hmac-sha2-256
-   `hmac-sha2-256-etm@openssh.com`


