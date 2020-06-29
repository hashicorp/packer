
// starts resources to provision them.
build {
    sources = [
        "source.virtualbox-iso.ubuntu-1204",
        "source.amazon-ebs.ubuntu-1604",
    ]

    post-processor "amazon-import" {
        only = ["virtualbox-iso.ubuntu-1204"]
    }
    post-processor "manifest" {
        except = ["virtualbox-iso.ubuntu-1204"]
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}

source "amazon-ebs" "ubuntu-1604" {
}
