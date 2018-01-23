---
description: |
    The ansible Packer provisioner allows Ansible playbooks to be run to
    provision the machine.
layout: docs
page_title: 'Ansible - Provisioners'
sidebar_current: 'docs-provisioners-ansible-remote'
---

# Ansible Provisioner

Type: `ansible`

The `ansible` Packer provisioner runs Ansible playbooks. It dynamically creates
an Ansible inventory file configured to use SSH, runs an SSH server, executes
`ansible-playbook`, and marshals Ansible plays through the SSH server to the
machine being provisioned by Packer. Note, this means that any `remote_user`
defined in tasks will be ignored. Packer will always connect with the user
given in the json config.

## Basic Example

This is a fully functional template that will provision an image on
DigitalOcean. Replace the mock `api_token` value with your own.

``` json
{
  "provisioners": [
    {
      "type": "ansible",
      "playbook_file": "./playbook.yml"
    }
  ],

  "builders": [
    {
      "type": "digitalocean",
      "api_token": "6a561151587389c7cf8faa2d83e94150a4202da0e2bad34dd2bf236018ffaeeb",
      "image": "ubuntu-14-04-x64",
      "region": "sfo1"
    }
  ]
}
```

## Configuration Reference

Required Parameters:

-   `playbook_file` - The playbook to be run by Ansible.

Optional Parameters:

-   `ansible_env_vars` (array of strings) - Environment variables to set before
    running Ansible.
    Usage example:

    ``` json
    {
      "ansible_env_vars": [ "ANSIBLE_HOST_KEY_CHECKING=False", "ANSIBLE_SSH_ARGS='-o ForwardAgent=yes -o ControlMaster=auto -o ControlPersist=60s'", "ANSIBLE_NOCOLOR=True" ]
    }
    ```

-   `command` (string) - The command to invoke ansible.
    Defaults to `ansible-playbook`.

-   `empty_groups` (array of strings) - The groups which should be present in
    inventory file but remain empty.

-   `extra_arguments` (array of strings) - Extra arguments to pass to Ansible.
    These arguments *will not* be passed through a shell and arguments should
    not be quoted. Usage example:

    ``` json
    {
      "extra_arguments": [ "--extra-vars", "Region={{user `Region`}} Stage={{user `Stage`}}" ]
    }
    ```

-   `groups` (array of strings) - The groups into which the Ansible host
    should be placed. When unspecified, the host is not associated with any
    groups.

-   `host_alias` (string) - The alias by which the Ansible host should be known.
    Defaults to `default`.

-   `inventory_directory` (string) - The directory in which to place the
    temporary generated Ansible inventory file. By default, this is the
    system-specific temporary file location. The fully-qualified name of this
    temporary file will be passed to the `-i` argument of the `ansible` command
    when this provisioner runs ansible. Specify this if you have an existing
    inventory directory with `host_vars` `group_vars` that you would like to use
    in the playbook that this provisioner will run.

-   `local_port` (string) - The port on which to attempt to listen for SSH
    connections. This value is a starting point. The provisioner will attempt
    listen for SSH connections on the first available of ten ports, starting at
    `local_port`. A system-chosen port is used when `local_port` is missing or
    empty.

-   `sftp_command` (string) - The command to run on the machine being provisioned
    by Packer to handle the SFTP protocol that Ansible will use to transfer
    files. The command should read and write on stdin and stdout, respectively.
    Defaults to `/usr/lib/sftp-server -e`.

-   `skip_version_check` (boolean) - Check if ansible is installed prior to running.
    Set this to `true`, for example, if you're going to install ansible during
    the packer run.

-   `ssh_host_key_file` (string) - The SSH key that will be used to run the SSH
    server on the host machine to forward commands to the target machine. Ansible
    connects to this server and will validate the identity of the server using
    the system known\_hosts. The default behavior is to generate and use a
    onetime key. Host key checking is disabled via the
    `ANSIBLE_HOST_KEY_CHECKING` environment variable if the key is generated.

-   `ssh_authorized_key_file` (string) - The SSH public key of the Ansible
    `ssh_user`. The default behavior is to generate and use a onetime key. If
    this key is generated, the corresponding private key is passed to
    `ansible-playbook` with the `--private-key` option.

-   `user` (string) - The `ansible_user` to use. Defaults to the user running
    packer.

## Default Extra Variables

In addition to being able to specify extra arguments using the
`extra_arguments` configuration, the provisioner automatically defines certain
commonly useful Ansible variables:

-   `packer_build_name` is set to the name of the build that Packer is running.
    This is most useful when Packer is making multiple builds and you want to
    distinguish them slightly when using a common playbook.

-   `packer_builder_type` is the type of the builder that was used to create the
    machine that the script is running on. This is useful if you want to run
    only certain parts of the playbook on systems built with certain builders.

## Debugging

To debug underlying issues with Ansible, add `"-vvvv"` to `"extra_arguments"` to enable verbose logging.

``` json
{
  "extra_arguments": [ "-vvvv" ]
}
```

## Limitations

### Redhat / CentOS

Redhat / CentOS builds have been known to fail with the following error due to `sftp_command`, which should be set to `/usr/libexec/openssh/sftp-server -e`:

``` text
==> virtualbox-ovf: starting sftp subsystem
    virtualbox-ovf: fatal: [default]: UNREACHABLE! => {"changed": false, "msg": "SSH Error: data could not be sent to the remote host. Make sure this host can be reached over ssh", "unreachable": true}
```

### chroot communicator

Building within a chroot (e.g. `amazon-chroot`) requires changing the Ansible connection to chroot.

``` json
{
  "builders": [
    {
      "type": "amazon-chroot",
      "mount_path": "/mnt/packer-amazon-chroot",
      "region": "us-east-1",
      "source_ami": "ami-123456"
    }
  ],
  "provisioners": [
    {
      "type": "ansible",
      "extra_arguments": [
        "--connection=chroot",
        "--inventory-file=/mnt/packer-amazon-chroot,"
      ],
      "playbook_file": "main.yml"
    }
  ]
}
```

### winrm communicator

Windows builds require a custom Ansible connection plugin and a particular configuration. Assuming a directory named `connection_plugins` is next to the playbook and contains a file named `packer.py` whose contents is

``` python
from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

from ansible.plugins.connection.ssh import Connection as SSHConnection

class Connection(SSHConnection):
    ''' ssh based connections for powershell via packer'''

    transport = 'packer'
    has_pipelining = True
    become_methods = []
    allow_executable = False
    module_implementation_preferences = ('.ps1', '')

    def __init__(self, *args, **kwargs):
        super(Connection, self).__init__(*args, **kwargs)
```

This template should build a Windows Server 2012 image on Google Cloud Platform:

``` json
{
  "variables": {},
  "provisioners": [
    {
      "type":  "ansible",
      "playbook_file": "./win-playbook.yml",
      "extra_arguments": [
        "--connection", "packer",
        "--extra-vars", "ansible_shell_type=powershell ansible_shell_executable=None"
      ]
    }
  ],
  "builders": [
    {
      "type": "googlecompute",
      "account_file": "{{user `account_file`}}",
      "project_id": "{{user `project_id`}}",
      "source_image": "windows-server-2012-r2-dc-v20160916",
      "communicator": "winrm",
      "zone": "us-central1-a",
      "disk_size": 50,
      "winrm_username": "packer",
      "winrm_use_ssl": true,
      "winrm_insecure": true,
      "metadata": {
        "sysprep-specialize-script-cmd": "winrm set winrm/config/service/auth @{Basic=\"true\"}"
      }
    }
  ]
}
```

### Too many SSH keys

SSH servers only allow you to attempt to authenticate a certain number of times. All of your loaded keys will be tried before the dynamically generated key. If you have too many SSH keys loaded in your `ssh-agent`, the Ansible provisioner may fail authentication with a message similar to this:

```console
    googlecompute: fatal: [default]: UNREACHABLE! => {"changed": false, "msg": "Failed to connect to the host via ssh: Warning: Permanently added '[127.0.0.1]:62684' (RSA) to the list of known hosts.\r\nReceived disconnect from 127.0.0.1 port 62684:2: too many authentication failures\r\nAuthentication failed.\r\n", "unreachable": true}
```

To unload all keys from your `ssh-agent`, run:

```console
$ ssh-add -D
```
