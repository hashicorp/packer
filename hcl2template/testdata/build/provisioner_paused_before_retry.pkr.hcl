
// starts resources to provision them.
build {
    sources = [
        "source.virtualbox-iso.ubuntu-1204"
    ]

    provisioner "shell" {
        pause_before = "10s"
        max_retries = 5
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}