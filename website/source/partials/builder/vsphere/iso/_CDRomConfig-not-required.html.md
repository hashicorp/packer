<!-- Code generated from the comments of the CDRomConfig struct in builder/vsphere/iso/step_add_cdrom.go; DO NOT EDIT MANUALLY -->

-   `cdrom_type` (string) - Which controller to use. Example: `sata`. Defaults to `ide`.
    
-   `remove_cdrom` (boolean) - Remove CD/DVD-ROM devices from template. Ddefaults to `false`.
    
-   `iso_paths` ([]string) - List of datastore paths to ISO files that will be mounted to the VM.
    Example: `"[datastore1] ISO/ubuntu.iso"`.
    
