source "null" "test" {}

build {
    sources = [ "null.test" ]

    post-processor "nonexistent" {
        foo = "bar"
    }
}
