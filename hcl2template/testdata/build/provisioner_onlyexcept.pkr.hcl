
// starts resources to provision them.
build {
    sources = [
        "source.virtualbox-iso.ubuntu-1204",
        "source.amazon-ebs.ubuntu-1604",
    ]

    provisioner "shell" {
        only = ["source.virtualbox-iso.ubuntu-1204"]
    }
    provisioner "file" {
        except = ["source.virtualbox-iso.ubuntu-1204"]
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}

source "amazon-ebs" "ubuntu-1604" {
}
