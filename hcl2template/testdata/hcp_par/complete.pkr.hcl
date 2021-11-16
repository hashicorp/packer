build {
    name = "bucket-slug"

    hcp_packer_registry {
        description = <<EOT
Some description
    EOT
        bucket_labels = {
            "foo" = "bar"
        }
        build_labels = {
            "python_version" = "3.0"
        }
    }

    sources = [
        "source.virtualbox-iso.ubuntu-1204",
    ]

    source "source.amazon-ebs.ubuntu-1604" {
      name = "aws-ubuntu-16.04"
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}

source "amazon-ebs" "ubuntu-1604" {
}
