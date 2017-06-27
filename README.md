# Packer Builder for VMware vSphere

This builder uses native vSphere API, and creates virtual machines remotely.

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

      "url":      "https://vcenter.domain.com/sdk",
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
* `url`
* `username`
* `password`
* `template`
* `vm_name`
* `host`
* `ssh_username`
* `ssh_password`

### Optional
Destination:
* `dc_name` (source datacenter)
* `resource_pool`
* `datastore`
* `linked_clone`

Hardware customization:
* `cpus`
* `ram`
* `shutdown_command`

Post-processing:
* `create_snapshot`
* `convert_to_template`

## Complete Example
```json
{
    "builders": [
        {
            "type": "vsphere",

            "url": "https://your.lab.addr/",
            "username": "username",
            "password": "secret",

            "ssh_username": "ssh_username",
            "ssh_password": "ssh_secret",

            "template": "template_name",
            "vm_name": "clone_name",
            "host": "172.16.0.1",
            "linked_clone": true,
            "create_snapshot": true,
            "convert_to_template": true,

            "RAM": "1024",
            "cpus": "2",
            "shutdown_command": "echo 'ssh_secret' | sudo -S shutdown -P now"
        } 
    ],
    "provisioners": [
        {
              "type": "shell",
              "inline": ["echo foo"]
        }
    ]
}
```
where `vm_name`, `RAM`, `cpus` and `shutdown_command` are parameters of the new VM. 
Parameters `ssh_*`, `dc_name` (datacenter name) and `template` (the name of the base VM) are for the base VM, 
on which you are creating the new one (note that VMWare Tools should be already installed on this template machine).
`vm_name` and `host` (describe the name of the new VM and the name of the host where we want to create it) are required parameters; you can also specify `resource_pool` (if you don't, the builder will try to detect the default one) and `datastore`.
`url`, `username` and `password` are your vSphere parameters.
