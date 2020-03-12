<!-- Code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; DO NOT EDIT MANUALLY -->

-   `name` (string) - name of the ovf. defaults to the name of the VM
    
-   `force` (bool) - overwrite ovf if it exists
    
-   `images` (bool) - include iso and img image files that are attached to the VM
    
-   `sha` (int) - generate manifest using SHA 1, 256, 512. use 0 (default) for no manifest
    
-   `options` ([]string) - Advanced ovf export options. Options can include:
    * mac - MAC address is exported for all ethernet devices
    * uuid - UUID is exported for all virtual machines
    * extraconfig - all extra configuration options are exported for a virtual machine
    * nodevicesubtypes - resource subtypes for CD/DVD drives, floppy drives, and serial and parallel ports are not exported
    
    For example, this config would output the mac addresses for all ethernet devices in the ovf file:
    ```json
    ...
      "export": {
        "options": ["mac"]
      },
    ```
    