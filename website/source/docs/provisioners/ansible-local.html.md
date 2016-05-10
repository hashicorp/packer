---
description: |
    The `ansible-local` Packer provisioner configures Ansible to run on the machine
    by Packer from local Playbook and Role files. Playbooks and Roles can be
    uploaded from your local machine to the remote machine. Ansible is run in local
    mode via the `ansible-playbook` command.
layout: docs
page_title: 'Ansible (Local) Provisioner'
...

# Ansible Local Provisioner

Type: `ansible-local`

The `ansible-local` Packer provisioner configures Ansible to run on the machine
by Packer from local Playbook and Role files. Playbooks and Roles can be
uploaded from your local machine to the remote machine. Ansible is run in [local
mode](https://docs.ansible.com/ansible/playbooks_delegation.html#local-playbooks) via the
`ansible-playbook` command.

-&gt; **Note:** Ansible will *not* be installed automatically by this
provisioner. This provisioner expects that Ansible is already installed on the
machine. It is common practice to use the [shell
provisioner](/docs/provisioners/shell.html) before the Ansible provisioner to do
this.

## Basic Example

The example below is fully functional.

``` {.javascript}
{
  "type": "ansible-local",
  "playbook_file": "local.yml"
}
```

## Configuration Reference

The reference of available configuration options is listed below.

Required:

-   `playbook_file` (string) - The playbook file to be executed by ansible. This
    file must exist on your local system and will be uploaded to the
    remote machine.

Optional:

-   `command` (string) - The command to invoke ansible. Defaults
    to "ansible-playbook".

-   `extra_arguments` (array of strings) - An array of extra arguments to pass
    to the ansible command. By default, this is empty.
    Usage example:

```
"extra_arguments": [ "--extra-vars \"Region={{user `Region`}} Stage={{user `Stage`}}\"" ]
```

-   `inventory_groups` (string) - A comma-separated list of groups to which
    packer will assign the host `127.0.0.1`. A value of `my_group_1,my_group_2`
    will generate an Ansible inventory like:

```{.text}
[my_group_1]
127.0.0.1
[my_group_2]
127.0.0.1
```

-   `inventory_file` (string) - The inventory file to be used by ansible. This
    file must exist on your local system and will be uploaded to the
    remote machine.

When using an inventory file, it's also required to `--limit` the hosts to the
specified host you're buiding. The `--limit` argument can be provided in the
`extra_arguments` option.

An example inventory file may look like:

```{.text}
[chi-dbservers]
db-01 ansible_connection=local
db-02 ansible_connection=local

[chi-appservers]
app-01 ansible_connection=local
app-02 ansible_connection=local

[chi:children]
chi-dbservers
chi-appservers

[dbservers:children]
chi-dbservers

[appservers:children]
chi-appservers
```

-   `playbook_dir` (string) - a path to the complete ansible directory structure
    on your local system to be copied to the remote machine as the
    `staging_directory` before all other files and directories.

-   `playbook_paths` (array of strings) - An array of paths to playbook files on
    your local system. These will be uploaded to the remote machine under
    `staging_directory`/playbooks. By default, this is empty.

-   `group_vars` (string) - a path to the directory containing ansible group
    variables on your local system to be copied to the remote machine. By
    default, this is empty.

-   `host_vars` (string) - a path to the directory containing ansible host
    variables on your local system to be copied to the remote machine. By
    default, this is empty.

-   `role_paths` (array of strings) - An array of paths to role directories on
    your local system. These will be uploaded to the remote machine under
    `staging_directory`/roles. By default, this is empty.

-   `staging_directory` (string) - The directory where all the configuration of
    Ansible by Packer will be placed. By default this
    is "/tmp/packer-provisioner-ansible-local". This directory doesn't need to
    exist but must have proper permissions so that the SSH user that Packer uses
    is able to create directories and write into this folder. If the permissions
    are not correct, use a shell provisioner prior to this to configure
    it properly.
