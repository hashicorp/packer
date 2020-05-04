
// starts resources to provision them.
build {
    sources = [
        "source.virtualbox-iso.ubuntu-1204"
    ]

    provisioner "shell" {
        timeout = "10s"
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}