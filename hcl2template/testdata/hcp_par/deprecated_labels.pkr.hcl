
build {

    hcp_packer_registry {
        bucket_name = "bucket-slug"
        labels = {
            "foo" = "bar"
        }
    }

    sources = [
        "source.virtualbox-iso.ubuntu-1204",
    ]
}

source "virtualbox-iso" "ubuntu-1204" {
}

