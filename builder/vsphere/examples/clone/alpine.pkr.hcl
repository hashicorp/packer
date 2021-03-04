# "timestamp" template function replacement
locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

# source blocks are analogous to the "builders" in json templates. They are used
# in build blocks. A build block runs provisioners and post-processors on a
# source. Read the documentation for source blocks here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/source
source "vsphere-clone" "example_clone" {
  communicator        = "none"
  host                = "esxi-1.vsphere65.test"
  insecure_connection = "true"
  password            = "jetbrains"
  template            = "alpine"
  username            = "root"
  vcenter_server      = "vcenter.vsphere65.test"
  vm_name             = "alpine-clone-${local.timestamp}"
}

# a build block invokes sources and runs provisioning steps on them. The
# documentation for build blocks can be found here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/build
build {
  sources = ["source.vsphere-clone.example_clone"]

}
