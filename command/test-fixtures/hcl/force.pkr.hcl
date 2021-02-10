source "null" "potato" {
    communicator = "none"
}

build {
    sources = ["sources.null.potato"]

    post-processor "manifest" {
        output = "manifest.json"
        strip_time = true
    }
}