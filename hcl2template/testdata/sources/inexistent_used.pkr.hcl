// a source represents a reusable setting for a system boot/start.
source "inexistant" "ubuntu-1204" {
    foo = "bar"
}

build {
    sources = ["inexistant.ubuntu-1204"]
}
