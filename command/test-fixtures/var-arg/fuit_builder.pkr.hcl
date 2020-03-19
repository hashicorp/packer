
variable "fruit" {
    type = string
}

source "null" "builder" {
    communicator = "none"
}

build {
    sources = [ 
        "source.null.builder",
    ]

    provisioner "shell-local" {
        inline = ["echo ${var.fruit} > ${var.fruit}.txt"]
    }
}
