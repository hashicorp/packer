<!-- Code generated from the comments of the Config struct in builder/virtualbox/iso/builder.go; DO NOT EDIT MANUALLY -->

-   `disk_size` (uint) - The size, in megabytes, of the hard disk to create for the VM. By
    default, this is 40000 (about 40 GB).
    
-   `guest_additions_mode` (string) - The method by which guest additions are made available to the guest for
    installation. Valid options are upload, attach, or disable. If the mode
    is attach the guest additions ISO will be attached as a CD device to the
    virtual machine. If the mode is upload the guest additions ISO will be
    uploaded to the path specified by guest_additions_path. The default
    value is upload. If disable is used, guest additions won't be
    downloaded, either.
    
-   `guest_additions_path` (string) - The path on the guest virtual machine where the VirtualBox guest
    additions ISO will be uploaded. By default this is
    VBoxGuestAdditions.iso which should upload into the login directory of
    the user. This is a configuration template where the `{{ .Version }}`
    variable is replaced with the VirtualBox version.
    
-   `guest_additions_sha256` (string) - The SHA256 checksum of the guest additions ISO that will be uploaded to
    the guest VM. By default the checksums will be downloaded from the
    VirtualBox website, so this only needs to be set if you want to be
    explicit about the checksum.
    
-   `guest_additions_url` (string) - The URL to the guest additions ISO to upload. This can also be a file
    URL if the ISO is at a local path. By default, the VirtualBox builder
    will attempt to find the guest additions ISO on the local file system.
    If it is not available locally, the builder will download the proper
    guest additions ISO from the internet. This is a template engine, and you
    have access to the variable `{{ .Version }}`.
    
-   `guest_additions_interface` (string) - The interface type to use to mount guest additions when
    guest_additions_mode is set to attach. Will default to the value set in
    iso_interface, if iso_interface is set. Will default to "ide", if
    iso_interface is not set. Options are "ide" and "sata".
    
-   `guest_os_type` (string) - The guest OS type being installed. By default this is other, but you can
    get dramatic performance improvements by setting this to the proper
    value. To view all available values for this run VBoxManage list
    ostypes. Setting the correct value hints to VirtualBox how to optimize
    the virtual hardware to work best with that operating system.
    
-   `hard_drive_discard` (bool) - When this value is set to true, a VDI image will be shrunk in response
    to the trim command from the guest OS. The size of the cleared area must
    be at least 1MB. Also set hard_drive_nonrotational to true to enable
    TRIM support.
    
-   `hard_drive_interface` (string) - The type of controller that the primary hard drive is attached to,
    defaults to ide. When set to sata, the drive is attached to an AHCI SATA
    controller. When set to scsi, the drive is attached to an LsiLogic SCSI
    controller. When set to pcie, the drive is attached to an NVMe
    controller. Please note that when you use "pcie", you'll need to have
    Virtualbox 6, install an [extension
    pack](https://www.virtualbox.org/wiki/Downloads#VirtualBox6.0.14OracleVMVirtualBoxExtensionPack)
    and you will need to enable EFI mode for nvme to work, ex:
    
    ```json
     "vboxmanage": [
          [ "modifyvm", "{{.Name}}", "--firmware", "EFI" ],
     ]
    ```
    
-   `sata_port_count` (int) - The number of ports available on any SATA controller created, defaults
    to 1. VirtualBox supports up to 30 ports on a maximum of 1 SATA
    controller. Increasing this value can be useful if you want to attach
    additional drives.
    
-   `nvme_port_count` (int) - The number of ports available on any NVMe controller created, defaults
    to 1. VirtualBox supports up to 255 ports on a maximum of 1 NVMe
    controller. Increasing this value can be useful if you want to attach
    additional drives.
    
-   `hard_drive_nonrotational` (bool) - Forces some guests (i.e. Windows 7+) to treat disks as SSDs and stops
    them from performing disk fragmentation. Also set hard_drive_discard to
    true to enable TRIM support.
    
-   `iso_interface` (string) - The type of controller that the ISO is attached to, defaults to ide.
    When set to sata, the drive is attached to an AHCI SATA controller.
    
-   `keep_registered` (bool) - Set this to true if you would like to keep the VM registered with
    virtualbox. Defaults to false.
    
-   `skip_export` (bool) - Defaults to false. When enabled, Packer will not export the VM. Useful
    if the build output is not the resultant image, but created inside the
    VM.
    
-   `vm_name` (string) - This is the name of the OVF file for the new virtual machine, without
    the file extension. By default this is packer-BUILDNAME, where
    "BUILDNAME" is the name of the build.
    