source "vmware-iso" "esxi-base-ubuntu-amd64" {
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

    remote_type             = "esx5"
    remote_host             = var.esxi_host
    remote_username         = var.esxi_user
    remote_password         = var.esxi_password
    remote_datastore        = var.datastore
    remote_private_key_file = var.remote_private_key_file_path
}

source "vmware-iso" "esxi-base-alpine-amd64" {
    headless                = var.headless
    boot_wait               = "10s"
    guest_os_type           = "otherLinux64Guest"
    floppy_files            = local.floppy_files_alpine_vsphere
    ssh_port                = 22
    ssh_timeout             = "10000s"
    ssh_username            = "root"
    ssh_password            = var.alpine_password
    tools_upload_flavor     = "linux"
    shutdown_command        = "poweroff"
    vmx_data = {
        "cpuid.coresPerSocket" = "1"
        "ethernet0.pciSlotNumber" = "32"
    }
    vmx_remove_ethernet_interfaces = true

    remote_type             = "esx5"
    remote_host             = var.esxi_host
    remote_username         = var.esxi_user
    remote_password         = var.esxi_password
    remote_datastore        = var.datastore
    remote_private_key_file = var.remote_private_key_file_path
}
