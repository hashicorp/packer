
// starts resources to provision them.
build {
    sources = [
        "source.virtualbox-iso.ubuntu-1204",
    ]

    provisioner "shell" {
        slice_string = ["{{packer_version}}"]
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}

