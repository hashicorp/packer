
source "parallels-iso" "base-ubuntu-amd64" {
  boot_wait              = "10s"
  guest_os_type          = "ubuntu"
  http_content           = local.http_directory_content
  parallels_tools_flavor = "lin"
  prlctl_version_file    = ".prlctl_version"
  shutdown_command       = "echo 'vagrant' | sudo -S shutdown -P now"
  ssh_password           = "vagrant"
  ssh_port               = 22
  ssh_timeout            = "10000s"
  ssh_username           = "vagrant"
}
