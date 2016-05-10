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
  the system known_hosts. The default behaviour is to generate and use a
  onetime key. Host key checking is disabled via the
  `ANSIBLE_HOST_KEY_CHECKING` environment variable if the key is generated.

- `ssh_authorized_key_file` (string) - The SSH public key of the Ansible
  `ssh_user`. The default behaviour is to generate and use a onetime key. If
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

- `extra_arguments` (array of strings) - Extra arguments to pass to Ansible.
  Usage example:

```
"extra_arguments": [ "--extra-vars", "Region={{user `Region`}} Stage={{user `Stage`}}" ]
```

- `ansible_env_vars` (array of strings) - Environment variables to set before running Ansible.
  If unset, defaults to `ANSIBLE_HOST_KEY_CHECKING=False`.
  Usage example:

```
"ansible_env_vars": [ "ANSIBLE_HOST_KEY_CHECKING=False", "ANSIBLE_SSH_ARGS='-o ForwardAgent=yes -o ControlMaster=auto -o ControlPersist=60s'", "ANSIBLE_NOCOLOR=True" ]
```

- `user` (string) - The `ansible_user` to use. Defaults to the user running
  packer.

## Limitations

The `ansible` provisioner does not support SCP to transfer files.
