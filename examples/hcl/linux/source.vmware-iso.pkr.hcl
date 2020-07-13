source "vmware-iso" "base-ubuntu-amd64" {
    headless                = var.headless
    boot_wait               = "10s"
    guest_os_type           = "ubuntu-64"
    http_directory          = local.http_directory
    shutdown_command        = "echo 'vagrant' | sudo -S shutdown -P now"
    ssh_password            = "vagrant"
    ssh_port                = 22
    ssh_timeout             = "10000s"
    ssh_username            = "vagrant"
    tools_upload_flavor     = "linux"
    vmx_data = {
        "cpuid.coresPerSocket" = "1"
        "ethernet0.pciSlotNumber" = "32"
    }
    vmx_remove_ethernet_interfaces = true
}
