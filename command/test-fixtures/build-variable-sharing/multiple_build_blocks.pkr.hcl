source "null" "chocolate" {
    communicator = "none"
}

source "null" "banana" {
    communicator = "none"
}

build {
    name = "vanilla"
    sources = ["null.chocolate"]

    provisioner "shell-local" {
        inline = ["echo hi > vanilla.chocolate.provisioner.${build.ID}.txt"]
    }

    post-processor "shell-local" {
        inline = ["echo hi > vanilla.chocolate.post-processor.${build.ID}.txt"]
    }
}

build {
    name = "apple"
    sources = ["null.chocolate"]

    provisioner "shell-local" {
        inline = ["echo hi > apple.chocolate.provisioner.${build.ID}.txt"]
    }

    post-processor "shell-local" {
        inline = ["echo hi > apple.chocolate.post-processor.${build.ID}.txt"]
    }
}

build {
    name = "sugar"
    sources = ["null.banana"]

    provisioner "shell-local" {
        inline = ["echo hi > sugar.banana.provisioner.${build.ID}.txt"]
    }

    post-processor "shell-local" {
        inline = ["echo hi > sugar.banana.post-processor.${build.ID}.txt"]
    }
}