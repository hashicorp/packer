<!-- Code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; DO NOT EDIT MANUALLY -->

-   `name` (string) - name of the ovf. defaults to the name of the VM
    
-   `force` (bool) - overwrite ovf if it exists
    
-   `images` (bool) - include iso and img image files that are attached to the VM
    
-   `sha` (int) - generate manifest using SHA 1, 256, 512. use 0 (default) for no manifest
    