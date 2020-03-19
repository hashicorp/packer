<!-- Code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; DO NOT EDIT MANUALLY -->

-   `name` (string) - name of the ovf. defaults to the name of the VM
    
-   `force` (bool) - overwrite ovf if it exists
    
-   `images` (bool) - include iso and img image files that are attached to the VM
    
-   `manifest` (string) - generate manifest using sha1, sha256, sha512. Defaults to 'sha256'. Use 'none' for no manifest.
    
-   `options` ([]string) - ```json
    ...
      "export": {
        "options": ["mac"]
      },
    ```
    