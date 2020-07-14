// I use the following config with direnv and set the following values:
// export PKR_VAR_bastion_host=""
// export PKR_VAR_bastion_user=""
// export PKR_VAR_datacenter_name=""
// export PKR_VAR_esxi_host=""
// export PKR_VAR_esxi_password=""
// export PKR_VAR_esxi_user=""
// export PKR_VAR_vcenter_endpoint=""
// export PKR_VAR_vcenter_password=""
// export PKR_VAR_vcenter_user=""
// export PKR_VAR_vm_ip=""
// export PKR_VAR_gateway_ip=""
// ...

variable "vcenter_endpoint" { type = string }
variable "vcenter_user"     { type = string }
variable "vcenter_password" { type = string }
variable "esxi_host"        { type = string }
variable "esxi_password"    { type = string }
variable "esxi_user"        { type = string }
variable "datacenter_name"  { type = string }
variable "vm_ip"            { type = string }
variable "gateway_ip"       { type = string }
variable "datastore" {
    default = "datastore1"
}
variable "remote_private_key_file_path" {
    type = string
}

source "vsphere-iso" "base-ubuntu-amd64" {
    vcenter_server      = var.vcenter_endpoint
    username            = var.vcenter_user
    password            = var.vcenter_password
    host                = var.esxi_host
    insecure_connection = true

    datacenter      = var.datacenter_name
    datastore       = "datastore1"

    ssh_password    = "vagrant"
    ssh_username    = "vagrant"

    CPUs            = 1
    RAM             = 512 * 2
    RAM_reserve_all = true
    
    disk_controller_type = "pvscsi"
    floppy_files = [
        "etc/http/preseed_hardcoded_ip.cfg"
    ]
    guest_os_type = "ubuntu64Guest"
    network_adapters {
        network      = "VM Network"
        network_card = "vmxnet3"
    }
    storage {
        disk_size             = 32768
        disk_thin_provisioned = true
    }

    boot_command = [
        "<enter><wait><f6><wait><esc><wait>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
        "<bs><bs><bs>",
        "/install/vmlinuz",
        " initrd=/install/initrd.gz",
        " priority=critical",
        " locale=en_US",
        " file=/media/preseed_hardcoded_ip.cfg",
        " netcfg/get_ipaddress=${var.vm_ip}",
        " netcfg/get_gateway=${var.gateway_ip}",
        "<enter>"
    ]
}


source "vsphere-iso" "base-alpine-amd64" {
    vcenter_server      = var.vcenter_endpoint
    username            = var.vcenter_user
    password            = var.vcenter_password
    host                = var.esxi_host
    insecure_connection = true

    datacenter      = var.datacenter_name
    datastore       = "datastore1"

    ssh_username            = "root"
    ssh_password            = var.alpine_password

    CPUs            = 1
    RAM             = 512 * 2
    RAM_reserve_all = true

    guest_os_type           = "otherLinux64Guest"

    floppy_files            = local.floppy_files_alpine_vsphere

    network_adapters {
        network      = "VM Network"
        network_card = "vmxnet3"
    }

    storage {
        disk_size             = 32768
        disk_thin_provisioned = true
    }
}
