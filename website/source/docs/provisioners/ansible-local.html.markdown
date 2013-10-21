---
layout: "docs"
page_title: "Ansible (Local) Provisioner"
---

# Ansible Local Provisioner

Type: `ansible-local`

The `ansible-local` provisioner configures Ansible to run on the machine by
Packer from local Playbook and Role files.  Playbooks and Roles can be uploaded
from your local machine to the remote machine.  Ansible is run in [local mode](http://www.ansibleworks.com/docs/playbooks2.html#local-playbooks) via the `ansible-playbook` command.

## Basic Example

The example below is fully functional.

<pre class="prettyprint">
{
    "type": "ansible-local",
    "playbook_file": "local.yml"
}
</pre>

## Configuration Reference

The reference of available configuration options is listed below.

Required:

* `playbook_file` (string) - The playbook file to be executed by ansible.
  This file must exist on your local system and will be uploaded to the
  remote machine.

Optional:

* `playbook_paths` (array of strings) - An array of paths to playbook files on
  your local system. These will be uploaded to the remote machine under
  `staging_directory`/playbooks. By default, this is empty.

* `role_paths` (array of strings) - An array of paths to role directories on
  your local system. These will be uploaded to the remote machine under
  `staging_directory`/roles. By default, this is empty.

* `staging_directory` (string) - The directory where all the configuration of
  Ansible by Packer will be placed. By default this is "/tmp/packer-provisioner-ansible-local".
  This directory doesn't need to exist but must have proper permissions so that
  the SSH user that Packer uses is able to create directories and write into
  this folder. If the permissions are not correct, use a shell provisioner prior
  to this to configure it properly.
