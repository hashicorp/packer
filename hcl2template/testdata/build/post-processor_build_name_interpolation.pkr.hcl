build {
    name = "test-build"
    sources = [ "source.virtualbox-iso.ubuntu-1204" ]

    post-processor "manifest" {
        name         = build.name
        slice_string = ["${packer.version}", "${build.name}"]
    }

}

source "virtualbox-iso" "ubuntu-1204" {
}
