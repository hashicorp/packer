source "null" "pizza" {
    communicator = "none"
}

build {
    name = "pineapple"
    sources = [
        "sources.null.pizza",
    ]

    provisioner "shell-local" {
        inline = [
            "echo '' > ${build.name}.${source.name}.txt"
        ]
    }
}
