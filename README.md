# packer-builder-vsphere

## The minimal working builder
``` json
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

You will find an example in **Installation instructions** section.

## Progress bar
You can find it [here](https://github.com/LizaTretyakova/packer-builder-vsphere/projects/1) as well.

- [x] hardware customization of the new VM (cpu, ram)
- [x] clone from template (not only from VM)
- [x] clone to alternate host, resource pool and datastore
- [ ] enable linked clones
- [ ] support Windows guest systems
- [ ] enable VM-to-template conversion
- [ ] tests
- [ ] add a shutdown timeout
- [ ] further hardware customization:
    * resize disks
    * ram reservation
    * cpu reservation

## Installation instructions

1. It is supposed that you already have Go(and [Packer](https://github.com/hashicorp/packer)), [Docker-compose](https://docs.docker.com/compose/install/) and [Glide](https://github.com/Masterminds/glide) set.

1. Download the sourcces from [github.com/LizaTretyakova/packer-builder-vsphere](github.com/LizaTretyakova/packer-builder-vsphere)

1. `cd` to `$GOPATH/go/src/github.com/LizaTretyakova/packer-builder-vsphere` (or wherever it was downloaded)

1. Get the dependencies
```
$ glide install
```

5. Build the binaries
```
$ docker-compose run build
```

6. The template for this builder is like following:
```json
{
    "builders": [
        {
            "type": "vsphere",
            "url": "https://your.url/",
            "username": "username",
            "password": "secret",
            "ssh_username": "ssh_username",
            "ssh_password": "ssh_secret",
            "dc_name": "datacenter1",
            "template": "template_vm_name",
            "vm_name": "new_vm_name",
            "host": "172.16.0.1",
            "resource_pool": "target_rpool",
            "datastore": "target_datastore",
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
`vm_name` and `host` (describe the name of the new VM and the name of the host where we want to create it) are required parameters; you can also specify `resource_pool` (if you don't, the builder will try to detect the default one) and `datastore` (**important**: if your target host differs from the initial one, you **have to** specify `datastore`; in case you stay within the same host, this parameter can be omitted). 
`url`, `username` and `password` are your vSphere parameters.
You need to set the appropriate values in the `variables` section before proceeding.

7. Now you can go to the `bin/` directory
```
$ cd ./bin
```
and try the builder
```
$ packer build template.json
```
