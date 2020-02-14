<!-- Code generated from the comments of the CreateConfig struct in builder/vsphere/iso/step_create.go; DO NOT EDIT MANUALLY -->

-   `vm_version` (uint) - Set VM hardware version. Defaults to the most current VM hardware
    version supported by vCenter. See
    [VMWare article 1003746](https://kb.vmware.com/s/article/1003746) for
    the full list of supported VM hardware versions.
    
-   `guest_os_type` (string) - Set VM OS type. Defaults to `otherGuest`. See [
    here](https://pubs.vmware.com/vsphere-6-5/index.jsp?topic=%2Fcom.vmware.wssdk.apiref.doc%2Fvim.vm.GuestOsDescriptor.GuestOsIdentifier.html)
    for a full list of possible values.
    
-   `firmware` (string) - Set the Firmware at machine creation. Example `efi`. Defaults to `bios`.
    
-   `disk_controller_type` (string) - Set VM disk controller type. Example `pvscsi`.
    
-   `disk_size` (int64) - The size of the disk in MB.
    
-   `disk_thin_provisioned` (bool) - Enable VMDK thin provisioning for VM. Defaults to `false`.
    
-   `network` (string) - Set network VM will be connected to.
    
-   `network_card` (string) - Set VM network card type. Example `vmxnet3`.
    
-   `network_adapters` ([]NIC) - Network adapters
    
-   `usb_controller` (bool) - Create USB controller for virtual machine. Defaults to `false`.
    
-   `notes` (string) - VM notes.
    