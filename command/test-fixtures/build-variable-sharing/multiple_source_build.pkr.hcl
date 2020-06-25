source "null" "chocolate" {
    communicator = "none"
}

source "null" "banana" {
    communicator = "none"
}

build {
    name = "vanilla"
    sources = [
        "null.chocolate",
        "null.banana",
    ]

    provisioner "shell-local" {
        inline = [
            "echo hi > all.${build.ID}.txt",
            "echo hi > chocolate.${build.ID}.txt",
            "echo hi > banana.${build.ID}.txt"
        ]
    }

    post-processor "shell-local" {
        only = ["null.chocolate"]
        inline = ["rm chocolate.${build.ID}.txt"]
    }
    post-processor "shell-local" {
        except = ["null.chocolate"]
        inline = ["rm banana.${build.ID}.txt"]
    }
}