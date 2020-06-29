
// starts resources to provision them.
build {
    sources = [
        "source.virtualbox-iso.ubuntu-1204",
    ]
    source "source.amazon-ebs.ubuntu-1604" {
        name = "aws-ubuntu-16.04"
    }

    post-processor "amazon-import" {
        only = ["virtualbox-iso.ubuntu-1204"]
    }
    post-processor "manifest" {
        except = ["virtualbox-iso.ubuntu-1204"]
    }
    post-processor "amazon-import" {
        only = ["amazon-ebs.aws-ubuntu-16.04"]
    }
    post-processor "manifest" {
        except = ["amazon-ebs.aws-ubuntu-16.04"]
    }
}

source "virtualbox-iso" "ubuntu-1204" {
}

source "amazon-ebs" "ubuntu-1604" {
}
