source "virtualbox-iso" "base-ubuntu-amd64" {
    headless                = var.headless
    guest_os_type           = "Ubuntu_64"
    http_directory          = local.http_directory
    shutdown_command        = "echo 'vagrant' | sudo -S shutdown -P now"
    ssh_username            = "vagrant"
    ssh_password            = "vagrant"
    ssh_port                = 22
    ssh_wait_timeout        = "15m"
    hard_drive_interface    = "sata"
    virtualbox_version_file = ".vbox_version"
    guest_additions_path    = "VBoxGuestAdditions_{{.Version}}.iso"
    guest_additions_url     = var.guest_additions_url
}
