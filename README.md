# Packer Builder for VMware vSphere

This a plugin for [HashiCorp Packer](https://www.packer.io/). It uses native vSphere API, and creates virtual machines remotely.

- VMware Player is not required
- Builds are incremental, VMs are not created from scratch but cloned from base templates - similar to [amazon-ebs](https://www.packer.io/docs/builders/amazon-ebs.html) builder
- Official vCenter API is used, no ESXi host [modification](https://www.packer.io/docs/builders/vmware-iso.html#building-on-a-remote-vsphere-hypervisor) is required 

## Usage
* Download the plugin from [Releases](https://github.com/jetbrains-infra/packer-builder-vsphere/releases) page
* [Install](https://www.packer.io/docs/extending/plugins.html#installing-plugins) the plugin, or simply put it into the same directory with configuration files

## Minimal Example

```json
{
  "builders": [
    {
      "type": "vsphere",

      "vcenter_server": "vcenter.domain.com",
      "username": "root",
      "password": "secret",

      "template": "ubuntu",
      "vm_name":  "vm-1",
      "host":     "esxi-1.domain.com",

      "ssh_username": "root",
      "ssh_password": "secret"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [ "echo hello" ]
    }
  ]
}
```

## Parameters
### Required
* `vcenter_server` - vCenter server hostname.
* `username` - vSphere username.
* `password` - vSphere password.
* `insecure_connection` - do not validate server's TLS certificate. `false` by default.

* `template` - name of source VM.
* `vm_name` - name of target VM.

* `host` - vSphere host where target VM is created.
* `ssh_username` - username in guest OS.
* `ssh_password` - password in guest OS.

### Optional
Destination:
* `datacenter` - required if there are several datacenters.
* `folder` - VM folder where target VM is created.
* `resource_pool` - by default a root of vSphere host.
* `datastore` - required if vSphere host has multiple datastores attached.
* `linked_clone` - create VM as a linked clone from latest snapshot. `false` by default.

Hardware customization:
* `CPUs` - number of CPU sockets. Inherited from source VM by default.
* `RAM` - Amount of RAM in megabytes. Inherited from source VM by default.

Post-processing:
* `shutdown_command` - VMware guest tools are used by default.
* `shutdown_timeout` - [Duration](https://golang.org/pkg/time/#ParseDuration) how long to wait for a graceful shutdown. 5 minutes by default.
* `create_snapshot` - add a snapshot, so VM can be used as a base for linked clones. `false` by default.
* `convert_to_template` - convert VM to a template. `false` by default.

## Complete Example
```json
{
  "variables": {
    "vsphere_password": "secret",
    "guest_password": "secret"
  },

  "builders": [
    {
      "type": "vsphere",

      "vcenter_server": "vcenter.domain.com",
      "username": "root",
      "password": "{{user `vsphere_password`}}",
      "insecure_connection": true,
      "datacenter": "dc1",

      "template": "ubuntu",
      "folder": "folder1/folder2",
      "vm_name": "vm-1",
      "host": "esxi-1.domain.com",
      "resource_pool": "pool1/pool2",
      "datastore": "datastore1",
      "linked_clone": true,

      "CPUs": 2,
      "RAM": 8192,

      "ssh_username": "root",
      "ssh_password": "{{user `guest_password`}}",

      "shutdown_command": "echo '{{user `guest_password`}}' | sudo -S shutdown -P now",
      "shutdown_timeout": "5m",
      "create_snapshot": true,
      "convert_to_template": true
    }
  ],

  "provisioners": [
    {
      "type": "shell",
      "environment_vars": [
        "DEBIAN_FRONTEND=noninteractive"
      ],
      "execute_command": "echo '{{user `guest_password`}}' | {{.Vars}} sudo -ES bash -eux '{{.Path}}'",
      "inline": [
        "apt-get install -y zip"
      ]
    }
  ]
}
```
