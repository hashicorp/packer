# source blocks are generated from your builders; a source can be referenced in
# build blocks. A build block runs provisioner and post-processors on a
# source. Read the documentation for source blocks here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/source
source "vsphere-iso" "example_windows" {
  CPUs                 = 1
  RAM                  = 4096
  RAM_reserve_all      = true
  communicator         = "winrm"
  disk_controller_type = ["pvscsi"]
  floppy_files         = ["${path.root}/setup/"]
  floppy_img_path      = "[datastore1] ISO/VMware Tools/10.2.0/pvscsi-Windows8.flp"
  guest_os_type        = "windows9_64Guest"
  host                 = "esxi-1.vsphere65.test"
  insecure_connection  = "true"
  iso_paths            = ["[datastore1] ISO/en_windows_10_multi-edition_vl_version_1709_updated_dec_2017_x64_dvd_100406172.iso", "[datastore1] ISO/VMware Tools/10.2.0/windows.iso"]
  network_adapters {
    network_card = "vmxnet3"
  }
  password = "jetbrains"
  storage {
    disk_size             = 32768
    disk_thin_provisioned = true
  }
  username       = "root"
  vcenter_server = "vcenter.vsphere65.test"
  vm_name        = "example-windows"
  winrm_password = "jetbrains"
  winrm_username = "jetbrains"
}

# a build block invokes sources and runs provisioning steps on them. The
# documentation for build blocks can be found here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/build
build {
  sources = ["source.vsphere-iso.example_windows"]

  provisioner "windows-shell" {
    inline = ["dir c:\\"]
  }
}
