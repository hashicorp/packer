<!-- Code generated from the comments of the Config struct in builder/vmware/iso/config.go; DO NOT EDIT MANUALLY -->

-   `disk_additional_size` ([]uint) - The size(s) of any additional
    hard disks for the VM in megabytes. If this is not specified then the VM
    will only contain a primary hard disk. The builder uses expandable, not
    fixed-size virtual hard disks, so the actual file representing the disk will
    not use the full size unless it is full.
    
-   `disk_adapter_type` (string) - The adapter type of the VMware virtual disk to create. This option is
    for advanced usage, modify only if you know what you're doing. Some of
    the options you can specify are `ide`, `sata`, `nvme` or `scsi` (which
    uses the "lsilogic" scsi interface by default). If you specify another
    option, Packer will assume that you're specifying a `scsi` interface of
    that specified type. For more information, please consult [Virtual Disk
    Manager User's Guide](http://www.vmware.com/pdf/VirtualDiskManager.pdf)
    for desktop VMware clients. For ESXi, refer to the proper ESXi
    documentation.
    
-   `vmdk_name` (string) - The filename of the virtual disk that'll be created,
    without the extension. This defaults to packer.
    
-   `disk_size` (uint) - The size of the hard disk for the VM in megabytes.
    The builder uses expandable, not fixed-size virtual hard disks, so the
    actual file representing the disk will not use the full size unless it
    is full. By default this is set to 40000 (about 40 GB).
    
-   `disk_type_id` (string) - The type of VMware virtual disk to create. This
    option is for advanced usage.
    
      For desktop VMware clients:
    
      Type ID | Description
      ------- | ---
      `0`     | Growable virtual disk contained in a single file (monolithic sparse).
      `1`     | Growable virtual disk split into 2GB files (split sparse).
      `2`     | Preallocated virtual disk contained in a single file (monolithic flat).
      `3`     | Preallocated virtual disk split into 2GB files (split flat).
      `4`     | Preallocated virtual disk compatible with ESX server (VMFS flat).
      `5`     | Compressed disk optimized for streaming.
    
      The default is `1`.
    
      For ESXi, this defaults to `zeroedthick`. The available options for ESXi
      are: `zeroedthick`, `eagerzeroedthick`, `thin`. `rdm:dev`, `rdmp:dev`,
      `2gbsparse` are not supported. Due to default disk compaction, when using
      `zeroedthick` or `eagerzeroedthick` set `skip_compaction` to `true`.
    
      For more information, please consult the [Virtual Disk Manager User's
      Guide](https://www.vmware.com/pdf/VirtualDiskManager.pdf) for desktop
      VMware clients. For ESXi, refer to the proper ESXi documentation.
    
-   `cdrom_adapter_type` (string) - The adapter type (or bus) that will be used
    by the cdrom device. This is chosen by default based on the disk adapter
    type. VMware tends to lean towards ide for the cdrom device unless
    sata is chosen for the disk adapter and so Packer attempts to mirror
    this logic. This field can be specified as either ide, sata, or scsi.
    
-   `guest_os_type` (string) - The guest OS type being installed. This will be
    set in the VMware VMX. By default this is other. By specifying a more
    specific OS type, VMware may perform some optimizations or virtual hardware
    changes to better support the operating system running in the
    virtual machine.
    
-   `version` (string) - The [vmx hardware
    version](http://kb.vmware.com/selfservice/microsites/search.do?language=en_US&cmd=displayKC&externalId=1003746)
    for the new virtual machine. Only the default value has been tested, any
    other value is experimental. Default value is `9`.
    
-   `vm_name` (string) - This is the name of the VMX file for the new virtual
    machine, without the file extension. By default this is packer-BUILDNAME,
    where "BUILDNAME" is the name of the build.
    
-   `vmx_disk_template_path` (string) - VMX Disk Template Path
-   `vmx_template_path` (string) - Path to a [configuration template](/docs/templates/engine.html) that
    defines the contents of the virtual machine VMX file for VMware. This is
    for **advanced users only** as this can render the virtual machine
    non-functional. See below for more information. For basic VMX
    modifications, try `vmx_data` first.
    