<!-- Code generated from the comments of the Config struct in builder/hyperv/vmcx/builder.go; DO NOT EDIT MANUALLY -->

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
    
-   `clone_from_vmcx_path` (string) - This is the path to a directory containing an exported virtual machine.
    
-   `clone_from_vm_name` (string) - This is the name of the virtual machine to clone from.
    
-   `clone_from_snapshot_name` (string) - The name of a snapshot in the
    source machine to use as a starting point for the clone. If the value
    given is an empty string, the last snapshot present in the source will
    be chosen as the starting point for the new VM.
    
-   `clone_all_snapshots` (bool) - If set to true all snapshots
    present in the source machine will be copied when the machine is
    cloned. The final result of the build will be an exported virtual
    machine that contains all the snapshots of the parent.
    
-   `vm_name` (string) - This is the name of the new virtual machine,
    without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.
    
-   `differencing_disk` (bool) - If true enables differencing disks. Only
    the changes will be written to the new disk. This is especially useful if
    your source is a VHD/VHDX. This defaults to false.
    
-   `switch_name` (string) - The name of the switch to connect the virtual
    machine to. By default, leaving this value unset will cause Packer to
    try and determine the switch to use by looking for an external switch
    that is up and running.
    
-   `copy_in_compare` (bool) - When cloning a vm to build from, we run a powershell
    Compare-VM command, which, depending on your version of Windows, may need
    the "Copy" flag to be set to true or false. Defaults to "false". Command:
    
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
-   `skip_compaction` (bool) - If true skip compacting the hard disk for
    the virtual machine when exporting. This defaults to false.
    
-   `skip_export` (bool) - If true Packer will skip the export of the VM.
    If you are interested only in the VHD/VHDX files, you can enable this
    option. The resulting VHD/VHDX file will be output to
    <output_directory>/Virtual Hard Disks. By default this option is false
    and Packer will export the VM to output_directory.
    
-   `headless` (bool) - Packer defaults to building Hyper-V virtual
    machines by launching a GUI that shows the console of the machine being
    built. When this value is set to true, the machine will start without a
    console.
    