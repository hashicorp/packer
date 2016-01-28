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

- `ssh_host_key_file` - The SSH key that will be used to run the SSH server to which Ansible connects.

- `ssh_authorized_key_file` - The SSH public key of the Ansible `ssh_user`.

Optional Parameters:

- `local_port` (string) - The port on which to 
  attempt to listen for SSH connections. This value is a starting point.
  The provisioner will attempt listen for SSH connections on the first
  available of ten ports, starting at `local_port`. The default value is 2200.

- `sftp_command` (string) - The command to run on the machine to handle the
  SFTP protocol that Ansible will use to transfer files. The command should
  read and write on stdin and stdout, respectively. Defaults to
  `/usr/lib/sftp-server -e`.

## Limitations

The `ansible` provisioner does not support SCP to transfer files.
