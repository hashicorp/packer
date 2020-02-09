<!-- Code generated from the comments of the Config struct in builder/vmware/vmx/config.go; DO NOT EDIT MANUALLY -->

-   `linked` (bool) - By default Packer creates a 'full' clone of the virtual machine
    specified in source_path. The resultant virtual machine is fully
    independant from the parent it was cloned from.
    
    Setting linked to true instead causes Packer to create the virtual
    machine as a 'linked' clone. Linked clones use and require ongoing
    access to the disks of the parent virtual machine. The benefit of a
    linked clone is that the clones virtual disk is typically very much
    smaller than would be the case for a full clone. Additionally, the
    cloned virtual machine can also be created much faster. Creating a
    linked clone will typically only be of benefit in some advanced build
    scenarios. Most users will wish to create a full clone instead. Defaults
    to false.
    
-   `vm_name` (string) - This is the name of the VMX file for the new virtual
    machine, without the file extension. By default this is packer-BUILDNAME,
    where "BUILDNAME" is the name of the build.
    