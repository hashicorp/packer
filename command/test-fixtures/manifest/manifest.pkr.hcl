source "file" "first-example" {
}

source "null" "second-example" {
    communicator = "none"
}

build {
    name = "a"

    source "sources.file.first-example" {
        name = "the_first_example"
        content = "cherry"
        target = "cherry.txt"
    }
    source "sources.file.first-example" {
        name = "the_second_example"
        content = "chocolate"
        target = "chocolate.txt"
    }

    post-processor "manifest" {
        output = "my.auto.pkrvars.hcl"
    }
}


build {
    sources = ["sources.null.second-example"]

        post-processor "manifest" {
        output = "my.auto.pkrvars.hcl"
    }
}
