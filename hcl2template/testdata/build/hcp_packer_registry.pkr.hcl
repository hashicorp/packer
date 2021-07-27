// starts resources to provision them.
build {
    hcp_packer_registry {
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
