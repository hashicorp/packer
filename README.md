# packer-builder-vsphere

## Usage
* Download the plugin from the [Releases](https://github.com/jetbrains-infra/packer-builder-vsphere/releases) page
* [Install](https://www.packer.io/docs/extending/plugins.html#installing-plugins) the plugin, or simply save it into the working directory together with a configuration file; you may create your own configuration file or take the one given below (remember to put your real values for names, passwords, `url` and `host`)
``` json
template.json

{
   "builders": [
      {
         "type": "vsphere",
         
         "url":          "https://your.lab.addr/",
         "username":     "username",
         "password":     "secret",
         
         "ssh_username": "ssh_username",
         "ssh_password": "ssh_secret",
         
         "template": "source_vm_name",
         "vm_name":  "clone_name",
         "host":     "172.16.0.1"
      }
   ]
}
```
(`host` is for target host)
* Run:
```
$ packer build template.json
```

## Builder parameters
### Required parameters:
* `username`
* `password`
* `template`
* `vm_name`
* `host`
### Optional parameters:
* Destination parameters:
    * `resource_pool`
    * `datastore`
* Hardware configuration:
    * `cpus`
    * `ram`
    * `shutdown_command`
* `ssh_username`
* `ssh_password`
* `dc_name` (source datacenter)
* Post-processing:
    * `linked_clone`
    * `create_snapshot`
    * `convert_to_template`

See an example below:
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


## Progress bar
You can find it [here](https://github.com/LizaTretyakova/packer-builder-vsphere/projects/1) as well.

- [x] hardware customization of the new VM (cpu, ram)
- [x] clone from template (not only from VM)
- [x] clone to alternate host, resource pool and datastore
- [x] enable linked clones
- [ ] support Windows guest systems
- [x] enable VM-to-template conversion
- [ ] tests
- [x] add a shutdown timeout
- [ ] further hardware customization:
    * resize disks
    * ram reservation
    * cpu reservation
