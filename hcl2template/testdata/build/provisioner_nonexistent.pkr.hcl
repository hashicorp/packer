source "null" "test" {}

build {
    sources = [ "null.test" ]

    provisioner "nonexistent" {
        foo = "bar"
    }
}
