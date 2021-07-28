build {
    name = "bucket-slug"

    hcp_packer_registry {
        slug = "real-bucket-slug"
        description = <<EOT
Some description
    EOT
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
