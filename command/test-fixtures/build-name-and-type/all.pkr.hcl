source "null" "test" {
    communicator = "none"
}

source "null" "potato" {
    communicator = "none"
}

build {
    sources = [
        "sources.null.test",
        "sources.null.potato",
    ]

    provisioner "shell-local" {
        inline = [
            "echo '' > ${source.type}.${source.name}.txt"
        ]
    }

    post-processor "manifest" {
        output = "manifest.json"
        strip_time = true
    }
}
