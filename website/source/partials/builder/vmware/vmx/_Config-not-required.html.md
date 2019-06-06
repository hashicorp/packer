<!-- Code generated from the comments of the Config struct in builder/vmware/vmx/config.go; DO NOT EDIT MANUALLY -->

-   `linked` (bool) - By default Packer creates a 'full' clone of
    the virtual machine specified in source_path. The resultant virtual
    machine is fully independant from the parent it was cloned from.
    
-   `remote_type` (string) - The type of remote machine that will be used to
    build this VM rather than a local desktop product. The only value accepted
    for this currently is esx5. If this is not set, a desktop product will
    be used. By default, this is not set.
    
-   `vm_name` (string) - This is the name of the VMX file for the new virtual
    machine, without the file extension. By default this is packer-BUILDNAME,
    where "BUILDNAME" is the name of the build.
    