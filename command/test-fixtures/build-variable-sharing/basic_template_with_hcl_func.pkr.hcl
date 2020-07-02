source "null" "chocolate" {
    communicator = "none"
}

build {
    sources = ["null.chocolate"]

    provisioner "shell-local" {
        inline = [
            "echo hi > provisioner.${build.ID}.txt",
            "echo hi > provisioner.${upper(build.ID)}.txt"
        ]
    }

    post-processor "shell-local" {
        inline = [
            "echo hi > post-processor.${build.ID}.txt",
            "echo hi > post-processor.${upper(build.ID)}.txt"
        ]
    }
}