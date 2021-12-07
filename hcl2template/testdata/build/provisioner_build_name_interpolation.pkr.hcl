
// starts resources to provision them.
build {
    name = "build-name-test"
    sources = [
        "source.virtualbox-iso.ubuntu-1204",
    ]

    provisioner "shell" {
        name = build.name
        slice_string = ["${build.name}"]
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}

