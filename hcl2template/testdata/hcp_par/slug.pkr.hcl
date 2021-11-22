build {
    name = "bucket-slug"

    hcp_packer_registry {
        bucket_name = "real-bucket-slug"
        description = <<EOT
Some description
    EOT
        bucket_labels = {
            "foo" = "bar"
        }
    }

    sources = [
        "source.virtualbox-iso.ubuntu-1204",
    ]
}

source "virtualbox-iso" "ubuntu-1204" {
}
