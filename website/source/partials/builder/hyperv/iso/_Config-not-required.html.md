<!-- Code generated from the comments of the Config struct in builder/hyperv/iso/builder.go; DO NOT EDIT MANUALLY -->

-   `disk_size` (uint) - The size, in megabytes, of the hard disk to create
    for the VM. By default, this is 40 GB.
    
-   `disk_block_size` (uint) - The block size of the VHD to be created.
    Recommended disk block size for Linux hyper-v guests is 1 MiB. This
    defaults to "32 MiB".
    
-   `memory` (uint) - The amount, in megabytes, of RAM to assign to the
    VM. By default, this is 1 GB.
    
-   `secondary_iso_images` ([]string) - A list of ISO paths to
    attach to a VM when it is booted. This is most useful for unattended
    Windows installs, which look for an Autounattend.xml file on removable
    media. By default, no secondary ISO will be attached.
    
-   `guest_additions_mode` (string) - If set to attach then attach and
    mount the ISO image specified in guest_additions_path. If set to
    none then guest additions are not attached and mounted; This is the
    default.
    
-   `guest_additions_path` (string) - The path to the ISO image for guest
    additions.
    
-   `vm_name` (string) - This is the name of the new virtual machine,
    without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.
    
-   `switch_name` (string) - The name of the switch to connect the virtual
    machine to. By default, leaving this value unset will cause Packer to
    try and determine the switch to use by looking for an external switch
    that is up and running.
    
-   `switch_vlan_id` (string) - This is the VLAN of the virtual switch's
    network card. By default none is set. If none is set then a VLAN is not
    set on the switch's network card. If this value is set it should match
    the VLAN specified in by vlan_id.
    
-   `mac_address` (string) - This allows a specific MAC address to be used on
    the default virtual network card. The MAC address must be a string with
    no delimiters, for example "0000deadbeef".
    
-   `vlan_id` (string) - This is the VLAN of the virtual machine's network
    card for the new virtual machine. By default none is set. If none is set
    then VLANs are not set on the virtual machine's network card.
    
-   `cpus` (uint) - The number of CPUs the virtual machine should use. If
    this isn't specified, the default is 1 CPU.
    
-   `generation` (uint) - The Hyper-V generation for the virtual machine. By
    default, this is 1. Generation 2 Hyper-V virtual machines do not support
    floppy drives. In this scenario use secondary_iso_images instead. Hard
    drives and DVD drives will also be SCSI and not IDE.
    
-   `enable_mac_spoofing` (bool) - If true enable MAC address spoofing
    for the virtual machine. This defaults to false.
    
-   `use_legacy_network_adapter` (bool) - If true use a legacy network adapter as the NIC.
    This defaults to false. A legacy network adapter is fully emulated NIC, and is thus
    supported by various exotic operating systems, but this emulation requires
    additional overhead and should only be used if absolutely necessary.
    
-   `enable_dynamic_memory` (bool) - If true enable dynamic memory for
    the virtual machine. This defaults to false.
    
-   `enable_secure_boot` (bool) - If true enable secure boot for the
    virtual machine. This defaults to false. See secure_boot_template
    below for additional settings.
    
-   `secure_boot_template` (string) - The secure boot template to be
    configured. Valid values are "MicrosoftWindows" (Windows) or
    "MicrosoftUEFICertificateAuthority" (Linux). This only takes effect if
    enable_secure_boot is set to "true". This defaults to "MicrosoftWindows".
    
-   `enable_virtualization_extensions` (bool) - If true enable
    virtualization extensions for the virtual machine. This defaults to
    false. For nested virtualization you need to enable MAC spoofing,
    disable dynamic memory and have at least 4GB of RAM assigned to the
    virtual machine.
    
-   `temp_path` (string) - The location under which Packer will create a
    directory to house all the VM files and folders during the build.
    By default %TEMP% is used which, for most systems, will evaluate to
    %USERPROFILE%/AppData/Local/Temp.
    
-   `configuration_version` (string) - This allows you to set the vm version when
     calling New-VM to generate the vm.
    
-   `keep_registered` (bool) - If "true", Packer will not delete the VM from
    The Hyper-V manager.
    
-   `communicator` (string) - Communicator
-   `disk_additional_size` ([]uint) - The size or sizes of any
    additional hard disks for the VM in megabytes. If this is not specified
    then the VM will only contain a primary hard disk. Additional drives
    will be attached to the SCSI interface only. The builder uses
    expandable rather than fixed-size virtual hard disks, so the actual
    file representing the disk will not use the full size unless it is
    full.
    
-   `skip_compaction` (bool) - If true skip compacting the hard disk for
    the virtual machine when exporting. This defaults to false.
    
-   `skip_export` (bool) - If true Packer will skip the export of the VM.
    If you are interested only in the VHD/VHDX files, you can enable this
    option. The resulting VHD/VHDX file will be output to
    <output_directory>/Virtual Hard Disks. By default this option is false
    and Packer will export the VM to output_directory.
    
-   `differencing_disk` (bool) - If true enables differencing disks. Only
    the changes will be written to the new disk. This is especially useful if
    your source is a VHD/VHDX. This defaults to false.
    
-   `use_fixed_vhd_format` (bool) - If true, creates the boot disk on the
    virtual machine as a fixed VHD format disk. The default is false, which
    creates a dynamic VHDX format disk. This option requires setting
    generation to 1, skip_compaction to true, and
    differencing_disk to false. Additionally, any value entered for
    disk_block_size will be ignored. The most likely use case for this
    option is outputing a disk that is in the format required for upload to
    Azure.
    
-   `headless` (bool) - Packer defaults to building Hyper-V virtual
    machines by launching a GUI that shows the console of the machine being
    built. When this value is set to true, the machine will start without a
    console.
    