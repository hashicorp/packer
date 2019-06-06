<!-- Code generated from the comments of the Config struct in builder/qemu/builder.go; DO NOT EDIT MANUALLY -->

-   `iso_skip_cache` (bool) - Use iso from provided url. Qemu must support
    curl block device. This defaults to false.
    
-   `accelerator` (string) - The accelerator type to use when running the VM.
    This may be none, kvm, tcg, hax, hvf, whpx, or xen. The appropriate
    software must have already been installed on your build machine to use the
    accelerator you specified. When no accelerator is specified, Packer will try
    to use kvm if it is available but will default to tcg otherwise.
    
-   `cpus` (int) - The number of cpus to use when building the VM.
     The default is 1 CPU.
    
-   `disk_interface` (string) - The interface to use for the disk. Allowed
    values include any of ide, scsi, virtio or virtio-scsi*. Note
    also that any boot commands or kickstart type scripts must have proper
    adjustments for resulting device names. The Qemu builder uses virtio by
    default.
    
-   `disk_size` (uint) - The size, in megabytes, of the hard disk to create
    for the VM. By default, this is 40960 (40 GB).
    
-   `disk_cache` (string) - The cache mode to use for disk. Allowed values
    include any of writethrough, writeback, none, unsafe
    or directsync. By default, this is set to writeback.
    
-   `disk_discard` (string) - The discard mode to use for disk. Allowed values
    include any of unmap or ignore. By default, this is set to ignore.
    
-   `disk_detect_zeroes` (string) - The detect-zeroes mode to use for disk.
    Allowed values include any of unmap, on or off. Defaults to off.
    When the value is "off" we don't set the flag in the qemu command, so that
    Packer still works with old versions of QEMU that don't have this option.
    
-   `skip_compaction` (bool) - Packer compacts the QCOW2 image using
    qemu-img convert.  Set this option to true to disable compacting.
    Defaults to false.
    
-   `disk_compression` (bool) - Apply compression to the QCOW2 disk file
    using qemu-img convert. Defaults to false.
    
-   `format` (string) - Either qcow2 or raw, this specifies the output
    format of the virtual machine image. This defaults to qcow2.
    
-   `headless` (bool) - Packer defaults to building QEMU virtual machines by
    launching a GUI that shows the console of the machine being built. When this
    value is set to true, the machine will start without a console.
    
-   `disk_image` (bool) - Packer defaults to building from an ISO file, this
    parameter controls whether the ISO URL supplied is actually a bootable
    QEMU image. When this value is set to true, the machine will either clone
    the source or use it as a backing file (if use_backing_file is true);
    then, it will resize the image according to disk_size and boot it.
    
-   `use_backing_file` (bool) - Only applicable when disk_image is true
    and format is qcow2, set this option to true to create a new QCOW2
    file that uses the file located at iso_url as a backing file. The new file
    will only contain blocks that have changed compared to the backing file, so
    enabling this option can significantly reduce disk usage.
    
-   `machine_type` (string) - The type of machine emulation to use. Run your
    qemu binary with the flags -machine help to list available types for
    your system. This defaults to pc.
    
-   `memory` (int) - The amount of memory to use when building the VM
    in megabytes. This defaults to 512 megabytes.
    
-   `net_device` (string) - The driver to use for the network interface. Allowed
    values ne2k_pci, i82551, i82557b, i82559er, rtl8139, e1000,
    pcnet, virtio, virtio-net, virtio-net-pci, usb-net, i82559a,
    i82559b, i82559c, i82550, i82562, i82557a, i82557c, i82801,
    vmxnet3, i82558a or i82558b. The Qemu builder uses virtio-net by
    default.
    
-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when packer
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is output-BUILDNAME where "BUILDNAME" is the
    name of the build.
    
-   `qemuargs` ([][]string) - Allows complete control over the
    qemu command line (though not, at this time, qemu-img). Each array of
    strings makes up a command line switch that overrides matching default
    switch/value pairs. Any value specified as an empty string is ignored. All
    values after the switch are concatenated with no separator.
    
-   `qemu_binary` (string) - The name of the Qemu binary to look for. This
    defaults to qemu-system-x86_64, but may need to be changed for
    some platforms. For example qemu-kvm, or qemu-system-i386 may be a
    better choice for some systems.
    
-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine unless a
    shutdown command takes place inside script so this may safely be omitted. It
    is important to add a shutdown_command. By default Packer halts the virtual
    machine and the file system may not be sync'd. Thus, changes made in a
    provisioner might not be saved. If one or more scripts require a reboot it is
    suggested to leave this blank since reboots may fail and specify the final
    shutdown command in your last script.
    
-   `ssh_host_port_min` (int) - The minimum and
    maximum port to use for the SSH port on the host machine which is forwarded
    to the SSH port on the guest machine. Because Packer often runs in parallel,
    Packer will choose a randomly available port in this range to use as the
    host port. By default this is 2222 to 4444.
    
-   `ssh_host_port_max` (int) - SSH Host Port Max
-   `use_default_display` (bool) - If true, do not pass a -display option
    to qemu, allowing it to choose the default. This may be needed when running
    under macOS, and getting errors about sdl not being available.
    
-   `vnc_bind_address` (string) - The IP address that should be
    binded to for VNC. By default packer will use 127.0.0.1 for this. If you
    wish to bind to all interfaces use 0.0.0.0.
    
-   `vnc_port_min` (int) - The minimum and maximum port
    to use for VNC access to the virtual machine. The builder uses VNC to type
    the initial boot_command. Because Packer generally runs in parallel,
    Packer uses a randomly chosen port in this range that appears available. By
    default this is 5900 to 6000. The minimum and maximum ports are inclusive.
    
-   `vnc_port_max` (int) - VNC Port Max
-   `vm_name` (string) - This is the name of the image (QCOW2 or IMG) file for
    the new virtual machine. By default this is packer-BUILDNAME, where
    "BUILDNAME" is the name of the build. Currently, no file extension will be
    used unless it is specified in this option.
    
-   `ssh_wait_timeout` (time.Duration) - These are deprecated, but we keep them around for BC
    TODO(@mitchellh): remove
    
-   `run_once` (bool) - TODO(mitchellh): deprecate
    
-   `shutdown_timeout` (string) - The amount of time to wait after executing the
    shutdown_command for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    5m or five minutes.
    