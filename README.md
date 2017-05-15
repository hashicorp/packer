# packer-builder-vsphere

## Installation instructions

0. It is supposed that you already have Go(and [Packer](https://github.com/hashicorp/packer)), [Docker-compose](https://docs.docker.com/compose/install/) and [Glide](https://github.com/Masterminds/glide) set.

1. Download the sourcces from [github.com/LizaTretyakova/packer-builder-vsphere](github.com/LizaTretyakova/packer-builder-vsphere)

2. `cd` to `$GOPATH/go/src/github.com/LizaTretyakova/packer-builder-vsphere` (or wherever it was downloaded)

3. Get the dependencies
```
$ glide install
```

4. Build the binaries
```
$ sudo docker-compose up
```

5. The template for this builder is like following:
```json
{
    "variables": {
            "url": "{{env `YOUR_VSPHERE_URL`}}",
            "username": "{{env `YOUR_VSPHERE_USERNAME`}}",
            "password": "{{env `YOUR_VSPHERE_PASSWORD`}}",
            "ssh_username": "{{env `TEMPLATE_VM_SSH_USERNAME`}}",
            "ssh_password": "{{env `TEMPLATE_VM_SSH_PASSWORD`}}",
            "dc_name": "{{env `TEMPLATE_VM_DATACENTER`}}",
            "template": "{{env `TEMPLATE_VM_NAME`}}"
    },
    "builders": [
        {
            "type": "vsphere",
            "url": "{{user `url`}}",
            "username": "{{user `username`}}",
            "password": "{{user `password`}}",
            "ssh_username": "{{user `ssh_username`}}",
            "ssh_password": "{{user `ssh_password`}}",
            "dc_name": "{{user `dc_name`}}",
            "template": "{{user `template`}}",
            "vm_name": "new_vm_name",
            "RAM": "1024",
            "cpus": "2",
            "shutdown_command": "echo '{{user `ssh_password`}}' | sudo -S shutdown -P now"
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
on which you are creating the new one (note, that VMWare Tools should be already installed on this template machine).
`url`, `username` and `password` are your vSphere parameters.
You need to set the appropriate values in the `variables` section before proceeding.

6. Now you can go to the `bin/` directory
```
$ cd ./bin
```
and try the builder
```
$ packer build template.json
```
