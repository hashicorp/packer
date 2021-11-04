source "null" "basic-example" {
}

build {
    sources = ["sources.null.basic-example"]

    provisioner "foo" {
        timeout = "10"
    }
}
