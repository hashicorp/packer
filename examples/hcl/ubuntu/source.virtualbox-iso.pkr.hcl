locals {
  // fileset lists all files in the http directory as a set, we convert that
  // set to a list of strings and we then take the directory of the first
  // value. This validates that the http directory exists even before starting
  // any builder/provisioner.
  http_directory = dirname(convert(fileset(".", "etc/http/*"), list(string))[0])
}

source "virtualbox-iso" "base-ubuntu-amd64" {
    boot_wait = "10s"
    guest_os_type = "Ubuntu_64"
    http_directory = local.http_directory
    shutdown_command = "echo 'vagrant' | sudo -S shutdown -P now"
    ssh_password = "vagrant"
    ssh_port = 22
    ssh_username = "vagrant"
    ssh_wait_timeout = "10000s"
}
