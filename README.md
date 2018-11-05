[![Team project](http://jb.gg/badges/team.svg)](https://confluence.jetbrains.com/display/ALL/JetBrains+on+GitHub)
[![Github All Releases](https://img.shields.io/github/downloads/jetbrains-infra/packer-builder-vsphere/total.svg)](https://github.com/jetbrains-infra/packer-builder-vsphere/releases)
[![TeamCity build status](https://img.shields.io/teamcity/http/teamcity.jetbrains.com/s/PackerVSphere_Build.svg)](https://teamcity.jetbrains.com/viewType.html?buildTypeId=PackerVSphere_Build&guest=1)


# Packer Builder for VMware vSphere

This a plugin for [HashiCorp Packer](https://www.packer.io/). It uses native vSphere API, and creates virtual machines remotely.

`vsphere-iso` builder creates new VMs from scratch.
`vsphere-clone` builder clones VMs from existing templates.

- VMware Player is not required.
- Official vCenter API is used, no ESXi host [modification](https://www.packer.io/docs/builders/vmware-iso.html#building-on-a-remote-vsphere-hypervisor) is required.

## Installation
* Download binaries from the [releases page](https://github.com/jetbrains-infra/packer-builder-vsphere/releases).
* [Install](https://www.packer.io/docs/extending/plugins.html#installing-plugins) the plugins, or simply put them into the same directory with JSON templates. On Linux and macOS run `chmod +x` on the files.

## Build

Install Go and [dep](https://github.com/golang/dep/releases), run `build.sh`.

Or build inside a container by Docker Compose:
```
docker-compose run build
```

The binaries will be in `bin/` directory.

Artifacts can be also downloaded from [TeamCity builds](https://teamcity.jetbrains.com/viewLog.html?buildTypeId=PackerVSphere_Build&buildId=lastSuccessful&tab=artifacts).

## Examples

See complete Ubuntu, Windows, and macOS templates in the [examples folder](https://github.com/jetbrains-infra/packer-builder-vsphere/tree/master/examples/).

## Parameter Reference

### Connection

* `vcenter_server`(string) - vCenter server hostname.
* `username`(string) - vSphere username.
* `password`(string) - vSphere password.
* `insecure_connection`(boolean) - Do not validate vCenter server's TLS certificate. Defaults to `false`.
* `datacenter`(string) - VMware datacenter name. Required if there is more than one datacenter in vCenter.

### VM Location

* `vm_name`(string) - Name of the new VM to create.
* `folder`(string) - VM folder to create the VM in.
* `host`(string) - ESXi host where target VM is created. A full path must be specified if the host is in a folder. For example `folder/host`. See the `Specifying Clusters and Hosts` section above for more details.
* `cluster`(string)  - ESXi cluster where target VM is created. See [Working with Clusters](#working-with-clusters) section.
* `resource_pool`(string) - VMWare resource pool. Defaults to the root resource pool of the `host` or `cluster`.
* `datastore`(string) - VMWare datastore. Required if `host` is a cluster, or if `host` has multiple datastores.

### VM Location (`vsphere-clone` only)

* `template`(string) - Name of source VM. Path is optional.
* `linked_clone`(boolean) - Create VM as a linked clone from latest snapshot. Defaults to `false`.

### Hardware

* `CPUs`(number) - Number of CPU sockets.
* `CPU_limit`(number) - Upper limit of available CPU resources in MHz.
* `CPU_reservation`(number) - Amount of reserved CPU resources in MHz.
* `CPU_hot_plug`(boolean) - Enable CPU hot plug setting for virtual machine. Defaults to `false`.
* `RAM`(number) - Amount of RAM in MB.
* `RAM_reservation`(number) - Amount of reserved RAM in MB.
* `RAM_reserve_all`(boolean) - Reserve all available RAM. Defaults to `false`. Cannot be used together with `RAM_reservation`.
* `RAM_hot_plug`(boolean) - Enable RAM hot plug setting for virtual machine. Defaults to `false`.
* `disk_size`(number) - The size of the disk in MB.
* `NestedHV`(boolean) - Enable nested hardware virtualization for VM. Defaults to `false`.
* `configuration_parameters`(map) - Custom parameters.
* `boot_order`(string) - Priority of boot devices. Defaults to `disk,cdrom`

### Hardware (`vsphere-iso` only)

* `vm_version`(number) - Set VM hardware version. Defaults to the most current VM hardware version supported by vCenter. See [VMWare article 1003746](https://kb.vmware.com/s/article/1003746) for the full list of supported VM hardware versions.
* `guest_os_type`(string) - Set VM OS type. Defaults to `otherGuest`. See [here](https://pubs.vmware.com/vsphere-6-5/index.jsp?topic=%2Fcom.vmware.wssdk.apiref.doc%2Fvim.vm.GuestOsDescriptor.GuestOsIdentifier.html) for a full list of possible values.
* `disk_controller_type`(string) - Set VM disk controller type. Example `pvscsi`.
* `disk_thin_provisioned`(boolean) - Enable VMDK thin provisioning for VM. Defaults to `false`.
* `network`(string) - Set network VM will be connected to.
* `network_card`(string) - Set VM network card type. Example `vmxnet3`.
* `usb_controller`(boolean) - Create USB controller for virtual machine. Defaults to `false`.
* `firmware`(string) - Set the Firmware at machine creation. Example `efi`. Defaults to `bios`

### Boot (`vsphere-iso` only)

* `boot_wait`(string) Amount of time to wait for the VM to boot. Examples 45s and 10m. Defaults to 10 seconds. See the Go Lang [ParseDuration](https://golang.org/pkg/time/#ParseDuration) documentation for full details.
* `boot_command`(array of strings) - List of commands to type when the VM is first booted. Used to initalize the operating system installer.
* `floppy_dirs`(array of strings) - Seems to not do anything useful yet. Not implemented.
* `floppy_files`(array of strings) - List of local files to be mounted to the VM floppy drive. Can be used to make Debian preseed or RHEL kickstart files available to the VM.
* `floppy_img_path`(string) - Data store path to a floppy image that will be mounted to the VM. Cannot be used with `floppy_files` or `floppy_dir` options. Example `[datastore1] ISO/VMware Tools/10.2.0/pvscsi-Windows8.flp`.
* `iso_paths`(array of strings) - List of data store paths to ISO files that will be mounted to the VM. Example `"[datastore1] ISO/ubuntu-16.04.3-server-amd64.iso"`.

### Provision

* `communicator` - `ssh` (default), `winrm`, or `none` (create/clone, customize hardware, but do not boot).

* `ssh_username`(string) - Username in guest OS.
* `ssh_password`(string) - Password to access guest OS. Only specify `ssh_password` or `ssh_private_key_file`, but not both.
* `ssh_private_key_file`(string) - Path to the SSH private key file to access guest OS. Only specify `ssh_password` or `ssh_private_key_file`, but not both.

* `winrm_username`(string) - Username in guest OS.
* `winrm_password`(string) - Password to access guest OS.

* `shutdown_command`(string) - Specify a VM guest shutdown command. VMware guest tools are used by default.
* `shutdown_timeout`(string) - Amount of time to wait for graceful VM shutdown. Examples 45s and 10m. Defaults to 5m(5 minutes). See the Go Lang [ParseDuration](https://golang.org/pkg/time/#ParseDuration) documentation for full details.

### Postprocessing

* `create_snapshot`(boolean) - Create a snapshot when set to `true`, so the VM can be used as a base for linked clones. Defaults to `false`.
* `convert_to_template`(boolean) - Convert VM to a template. Defaults to `false`.

## Working with Clusters
#### Standalone Hosts
Only use the `host` option. Optionally specify a `resource_pool`:
```
"host": "esxi-1.vsphere65.test",
"resource_pool": "pool1",
```

#### Clusters Without DRS
Use the `cluster` and `host `parameters:
```
"cluster": "cluster1",
"host": "esxi-2.vsphere65.test",
```

#### Clusters With DRS
Only use the `cluster` option. Optionally specify a `resource_pool`:
```
"cluster": "cluster2",
"resource_pool": "pool1",
```

## Required vSphere Permissions

* VM folder (this object and children):
    ```
    Virtual machine -> Inventory
    Virtual machine -> Configuration
    Virtual machine -> Interaction
    Virtual machine -> Snapshot management
    Virtual machine -> Provisioning
    ```
    Individual privileges are listed in https://github.com/jetbrains-infra/packer-builder-vsphere/issues/97#issuecomment-436063235.
* Resource pool, host, or cluster (this object): 
    ```
    Resource -> Assign virtual machine to resource pool
    ```
* Host in clusters without DRS (this object): 
    ```
    Read-only
    ```
* Datastore (this object): 
    ```
    Datastore -> Allocate space
    Datastore -> Browse datastore
    Datastore -> Low level file operations
    ```  
* Network (this object): 
    ```
    Network -> Assign network
    ```
* Distributed switch (this object): 
    ```
    Read-only
    ```

For floppy image upload:

* Datacenter (this object): 
    ```
    Datastore -> Low level file operations
    ```
* Host (this object): 
    ```
    Host -> Configuration -> System Management
    ```
