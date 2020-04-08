<!-- Code generated from the comments of the Config struct in builder/qemu/builder.go; DO NOT EDIT MANUALLY -->

-   `iso_skip_cache` (bool) - Use iso from provided url. Qemu must support
    curl block device. This defaults to `false`.
    
-   `accelerator` (string) - The accelerator type to use when running the VM.
    This may be `none`, `kvm`, `tcg`, `hax`, `hvf`, `whpx`, or `xen`. The appropriate
    software must have already been installed on your build machine to use the
    accelerator you specified. When no accelerator is specified, Packer will try
    to use `kvm` if it is available but will default to `tcg` otherwise.
    
    -&gt; The `hax` accelerator has issues attaching CDROM ISOs. This is an
    upstream issue which can be tracked
    [here](https://github.com/intel/haxm/issues/20).
    
    -&gt; The `hvf` and `whpx` accelerator are new and experimental as of
    [QEMU 2.12.0](https://wiki.qemu.org/ChangeLog/2.12#Host_support).
    You may encounter issues unrelated to Packer when using these.  You may need to
    add [ "-global", "virtio-pci.disable-modern=on" ] to `qemuargs` depending on the
    guest operating system.
    
    -&gt; For `whpx`, note that [Stefan Weil's QEMU for Windows distribution](https://qemu.weilnetz.de/w64/)
    does not include WHPX support and users may need to compile or source a
    build of QEMU for Windows themselves with WHPX support.
    
-   `disk_additional_size` ([]string) - Additional disks to create. Uses `vm_name` as the disk name template and
    appends `-#` where `#` is the position in the array. `#` starts at 1 since 0
    is the default disk. Each string represents the disk image size in bytes.
    Optional suffixes 'k' or 'K' (kilobyte, 1024), 'M' (megabyte, 1024k), 'G'
    (gigabyte, 1024M), 'T' (terabyte, 1024G), 'P' (petabyte, 1024T) and 'E'
    (exabyte, 1024P)  are supported. 'b' is ignored. Per qemu-img documentation.
    Each additional disk uses the same disk parameters as the default disk.
    Unset by default.
    
-   `cpus` (int) - The number of cpus to use when building the VM.
     The default is `1` CPU.
    
-   `disk_interface` (string) - The interface to use for the disk. Allowed values include any of `ide`,
    `scsi`, `virtio` or `virtio-scsi`^\*. Note also that any boot commands
    or kickstart type scripts must have proper adjustments for resulting
    device names. The Qemu builder uses `virtio` by default.
    
    ^\* Please be aware that use of the `scsi` disk interface has been
    disabled by Red Hat due to a bug described
    [here](https://bugzilla.redhat.com/show_bug.cgi?id=1019220). If you are
    running Qemu on RHEL or a RHEL variant such as CentOS, you *must* choose
    one of the other listed interfaces. Using the `scsi` interface under
    these circumstances will cause the build to fail.
    
-   `disk_size` (string) - The size in bytes of the hard disk of the VM. Suffix with the first
    letter of common byte types. Use "k" or "K" for kilobytes, "M" for
    megabytes, G for gigabytes, and T for terabytes. If no value is provided
    for disk_size, Packer uses a default of `40960M` (40 GB). If a disk_size
    number is provided with no units, Packer will default to Megabytes.
    
-   `disk_cache` (string) - The cache mode to use for disk. Allowed values include any of
    `writethrough`, `writeback`, `none`, `unsafe` or `directsync`. By
    default, this is set to `writeback`.
    
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
    
-   `format` (string) - Either `qcow2` or `raw`, this specifies the output format of the virtual
    machine image. This defaults to `qcow2`.
    
-   `headless` (bool) - Packer defaults to building QEMU virtual machines by
    launching a GUI that shows the console of the machine being built. When this
    value is set to `true`, the machine will start without a console.
    
    You can still see the console if you make a note of the VNC display
    number chosen, and then connect using `vncviewer -Shared <host>:<display>`
    
-   `disk_image` (bool) - Packer defaults to building from an ISO file, this parameter controls
    whether the ISO URL supplied is actually a bootable QEMU image. When
    this value is set to `true`, the machine will either clone the source or
    use it as a backing file (if `use_backing_file` is `true`); then, it
    will resize the image according to `disk_size` and boot it.
    
-   `use_backing_file` (bool) - Only applicable when disk_image is true
    and format is qcow2, set this option to true to create a new QCOW2
    file that uses the file located at iso_url as a backing file. The new file
    will only contain blocks that have changed compared to the backing file, so
    enabling this option can significantly reduce disk usage.
    
-   `machine_type` (string) - The type of machine emulation to use. Run your qemu binary with the
    flags `-machine help` to list available types for your system. This
    defaults to `pc`.
    
-   `memory` (int) - The amount of memory to use when building the VM
    in megabytes. This defaults to 512 megabytes.
    
-   `net_device` (string) - The driver to use for the network interface. Allowed values `ne2k_pci`,
    `i82551`, `i82557b`, `i82559er`, `rtl8139`, `e1000`, `pcnet`, `virtio`,
    `virtio-net`, `virtio-net-pci`, `usb-net`, `i82559a`, `i82559b`,
    `i82559c`, `i82550`, `i82562`, `i82557a`, `i82557c`, `i82801`,
    `vmxnet3`, `i82558a` or `i82558b`. The Qemu builder uses `virtio-net` by
    default.
    
-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when packer
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is output-BUILDNAME where "BUILDNAME" is the
    name of the build.
    
-   `qemuargs` ([][]string) - Allows complete control over the qemu command line (though not, at this
    time, qemu-img). Each array of strings makes up a command line switch
    that overrides matching default switch/value pairs. Any value specified
    as an empty string is ignored. All values after the switch are
    concatenated with no separator.
    
    ~&gt; **Warning:** The qemu command line allows extreme flexibility, so
    beware of conflicting arguments causing failures of your run. For
    instance, using --no-acpi could break the ability to send power signal
    type commands (e.g., shutdown -P now) to the virtual machine, thus
    preventing proper shutdown. To see the defaults, look in the packer.log
    file and search for the qemu-system-x86 command. The arguments are all
    printed for review.
    
    The following shows a sample usage:
    
    ```json
    {
      "qemuargs": [
        [ "-m", "1024M" ],
        [ "--no-acpi", "" ],
        [
          "-netdev",
          "user,id=mynet0,",
          "hostfwd=hostip:hostport-guestip:guestport",
          ""
        ],
        [ "-device", "virtio-net,netdev=mynet0" ]
      ]
    }
    ```
    
    would produce the following (not including other defaults supplied by
    the builder and not otherwise conflicting with the qemuargs):
    
    ```text
    qemu-system-x86 -m 1024m --no-acpi -netdev
    user,id=mynet0,hostfwd=hostip:hostport-guestip:guestport -device
    virtio-net,netdev=mynet0"
    ```
    
    ~&gt; **Windows Users:** [QEMU for Windows](https://qemu.weilnetz.de/)
    builds are available though an environmental variable does need to be
    set for QEMU for Windows to redirect stdout to the console instead of
    stdout.txt.
    
    The following shows the environment variable that needs to be set for
    Windows QEMU support:
    
    ```text
    setx SDL_STDIO_REDIRECT=0
    ```
    
    You can also use the `SSHHostPort` template variable to produce a packer
    template that can be invoked by `make` in parallel:
    
    ```json
    {
      "qemuargs": [
        [ "-netdev", "user,hostfwd=tcp::{{ .SSHHostPort }}-:22,id=forward"],
        [ "-device", "virtio-net,netdev=forward,id=net0"]
      ]
    }
    ```
    
    `make -j 3 my-awesome-packer-templates` spawns 3 packer processes, each
    of which will bind to their own SSH port as determined by each process.
    This will also work with WinRM, just change the port forward in
    `qemuargs` to map to WinRM's default port of `5985` or whatever value
    you have the service set to listen on.
    
    This is a template engine and allows access to the following variables:
    `{{ .HTTPIP }}`, `{{ .HTTPPort }}`, `{{ .HTTPDir }}`,
    `{{ .OutputDir }}`, `{{ .Name }}`, and `{{ .SSHHostPort }}`
    
-   `qemu_binary` (string) - The name of the Qemu binary to look for. This
    defaults to qemu-system-x86_64, but may need to be changed for
    some platforms. For example qemu-kvm, or qemu-system-i386 may be a
    better choice for some systems.
    
-   `qmp_enable` (bool) - Enable QMP socket. Location is specified by `qmp_socket_path`. Defaults
    to false.
    
-   `qmp_socket_path` (string) - QMP Socket Path when `qmp_enable` is true. Defaults to
    `output_directory`/`vm_name`.monitor.
    
-   `ssh_host_port_min` (int) - The minimum and maximum port to use for the SSH port on the host machine
    which is forwarded to the SSH port on the guest machine. Because Packer
    often runs in parallel, Packer will choose a randomly available port in
    this range to use as the host port. By default this is 2222 to 4444.
    
-   `ssh_host_port_max` (int) - SSH Host Port Max
-   `use_default_display` (bool) - If true, do not pass a -display option
    to qemu, allowing it to choose the default. This may be needed when running
    under macOS, and getting errors about sdl not being available.
    
-   `display` (string) - What QEMU -display option to use. Defaults to gtk, use none to not pass the
    -display option allowing QEMU to choose the default. This may be needed when
    running under macOS, and getting errors about sdl not being available.
    
-   `vnc_bind_address` (string) - The IP address that should be
    binded to for VNC. By default packer will use 127.0.0.1 for this. If you
    wish to bind to all interfaces use 0.0.0.0.
    
-   `vnc_use_password` (bool) - Whether or not to set a password on the VNC server. This option
    automatically enables the QMP socket. See `qmp_socket_path`. Defaults to
    `false`.
    
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
    