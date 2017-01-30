---
layout: "docs"
page_title: "Ansible Provisioner"
description: |-
  The `ansible` Packer provisioner allows Ansible playbooks to be run to provision the machine.
---

# Ansible Provisioner

Type: `ansible`

The `ansible` Packer provisioner runs Ansible playbooks. It dynamically creates
an Ansible inventory file configured to use SSH, runs an SSH server, executes
`ansible-playbook`, and marshals Ansible plays through the SSH server to the
machine being provisioned by Packer.

## Basic Example

This is a fully functional template that will provision an image on
DigitalOcean. Replace the mock `api_token` value with your own.

```json
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

- `playbook_file` - The playbook to be run by Ansible.

Optional Parameters:

- `command` (string) - The command to invoke ansible.
   Defaults to `ansible-playbook`.

- `groups` (array of strings) - The groups into which the Ansible host
  should be placed. When unspecified, the host is not associated with any
  groups.

- `empty_groups` (array of strings) - The groups which should be present in
  inventory file but remain empty.

- `host_alias` (string) - The alias by which the Ansible host should be known.
  Defaults to `default`.

- `ssh_host_key_file` (string) - The SSH key that will be used to run the SSH
  server on the host machine to forward commands to the target machine. Ansible
  connects to this server and will validate the identity of the server using
  the system known_hosts. The default behavior is to generate and use a
  onetime key. Host key checking is disabled via the
  `ANSIBLE_HOST_KEY_CHECKING` environment variable if the key is generated.

- `ssh_authorized_key_file` (string) - The SSH public key of the Ansible
  `ssh_user`. The default behavior is to generate and use a onetime key. If
  this key is generated, the corresponding private key is passed to
  `ansible-playbook` with the `--private-key` option.

- `local_port` (string) - The port on which to attempt to listen for SSH
  connections. This value is a starting point.  The provisioner will attempt
  listen for SSH connections on the first available of ten ports, starting at
  `local_port`. A system-chosen port is used when `local_port` is missing or
  empty.

- `sftp_command` (string) - The command to run on the machine being provisioned
  by Packer to handle the SFTP protocol that Ansible will use to transfer
  files. The command should read and write on stdin and stdout, respectively.
  Defaults to `/usr/lib/sftp-server -e`.

- `use_sftp` (boolean) - Whether to use SFTP. When false,
  `ANSIBLE_SCP_IF_SSH=True` will be automatically added to `ansible_env_vars`.
  Defaults to false.

- `extra_arguments` (array of strings) - Extra arguments to pass to Ansible.
  Usage example:

```
"extra_arguments": [ "--extra-vars", "Region={{user `Region`}} Stage={{user `Stage`}}" ]
```

- `ansible_env_vars` (array of strings) - Environment variables to set before
  running Ansible.
  Usage example:

```
"ansible_env_vars": [ "ANSIBLE_HOST_KEY_CHECKING=False", "ANSIBLE_SSH_ARGS='-o ForwardAgent=yes -o ControlMaster=auto -o ControlPersist=60s'", "ANSIBLE_NOCOLOR=True" ]
```

- `user` (string) - The `ansible_user` to use. Defaults to the user running
  packer.

## Limitations

### Redhat / CentOS

Redhat / CentOS builds have been known to fail with the following error due to `sftp_command`, which should be set to `/usr/libexec/openssh/sftp-server -e`:

```
==> virtualbox-ovf: starting sftp subsystem
    virtualbox-ovf: fatal: [default]: UNREACHABLE! => {"changed": false, "msg": "SSH Error: data could not be sent to the remote host. Make sure this host can be reached over ssh", "unreachable": true}
```

### chroot communicator

Building within a chroot (e.g. `amazon-chroot`) requires changing the Ansible connection to chroot.

```
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

```
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

```
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
