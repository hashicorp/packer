---
layout: "docs"
page_title: "Ansible Provisioner"
description: |-
  The `ansible` Packer provisioner allows Ansible playbooks to be run to provision the machine. 
---

# Ansible Provisioner

Type: `ansible`

The `ansible` Packer provisioner allows Ansible playbooks to be run to provision the machine.

## Basic Example

This is a fully functional template that will provision an image on
DigitalOcean. Replace the mock `api_token` value with your own.

```json
{
  "provisioners": [
    {
      "type": "ansible",
      "playbook_file": "./playbook.yml",
      "extra_arguments": ["--private-key", "./id_packer-ansible", "-v", "-c", "paramiko"],
      "ssh_authorized_key_file": "./id_packer-ansible.pub",
      "ssh_host_key_file": "./packer_host_private_key"
    }
  ],

  "builders": [
    {
      "type": "digitalocean",
      "api_token": "6a561151587389c7cf8faa2d83e94150a4202da0e2bad34dd2bf236018ffaeeb",
      "image": "ubuntu-14-04-x64",
      "region": "sfo1"
    },
  ]
}
```

## Configuration Reference

Required Parameters:

- `playbook_file` - The playbook file to be run by Ansible.

Optional Parameters:

- `ssh_host_key_file` (string) - The SSH key that will be used to run the SSH
  server on the host machine to forward commands to the target machine. Ansible
  connects to this server and will validate the identity of the server using
  the system known_hosts. The default behaviour is to generate and use a
  onetime key, and disable host_key_verification in Ansible to allow it to
  connect to the server.

- `ssh_authorized_key_file` (string) - The SSH public key of the Ansible
  `ssh_user`. The default behaviour is to generate and use a onetime key. If
  this file is generated, the corresponding private key will be passed via the
  `--private-key` option to Ansible.

- `local_port` (string) - The port on which to attempt to listen for SSH
  connections. This value is a starting point.  The provisioner will attempt
  listen for SSH connections on the first available of ten ports, starting at
  `local_port`. When `local_port` is missing or empty, ansible-provisioner will
  listen on a system-chosen port.


- `sftp_command` (string) - The command to run on the provisioned machine to
  handle the SFTP protocol that Ansible will use to transfer files. The command
  should read and write on stdin and stdout, respectively. Defaults to
  `/usr/lib/sftp-server -e`.

- `extra_arguments` (string) - Extra arguments to pass to Ansible.

## Limitations

The `ansible` provisioner does not support SCP to transfer files.
