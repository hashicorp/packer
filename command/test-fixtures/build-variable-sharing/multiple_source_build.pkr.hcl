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
            "echo hi > all-provisioner.${build.ID}.txt",
            "echo hi > chocolate-provisioner.${build.ID}.txt",
            "echo hi > banana-provisioner.${build.ID}.txt"
        ]
    }
    provisioner "shell-local" {
        only = ["null.chocolate"]
        inline = ["rm chocolate-provisioner.${build.ID}.txt"]
    }
    provisioner "shell-local" {
        except = ["null.chocolate"]
        inline = ["rm banana-provisioner.${build.ID}.txt"]
    }

    post-processor "shell-local" {
        inline = [
            "echo hi > all-post-processor.${build.ID}.txt",
            "echo hi > chocolate-post-processor.${build.ID}.txt",
            "echo hi > banana-post-processor.${build.ID}.txt"
        ]
    }
    post-processor "shell-local" {
        only = ["null.chocolate"]
        inline = ["rm chocolate-post-processor.${build.ID}.txt"]
    }
    post-processor "shell-local" {
        except = ["null.chocolate"]
        inline = ["rm banana-post-processor.${build.ID}.txt"]
    }
}