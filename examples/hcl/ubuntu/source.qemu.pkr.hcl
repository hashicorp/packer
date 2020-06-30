source "qemu" "base-ubuntu-amd64" {
    headless         = var.headless
    floppy_files     = [
        "${local.http_directory}/preseed.cfg",
    ]
    http_directory   = local.http_directory
    shutdown_command = "echo 'vagrant'|sudo -S shutdown -P now"
    ssh_password     = "vagrant"
    ssh_username     = "vagrant"
    ssh_wait_timeout = "50m"
    disk_size        = 5000
    disk_interface   = "virtio-scsi"
    memory           = 512 * 4
    cpus             = 4
    boot_wait        = "5s"
}
