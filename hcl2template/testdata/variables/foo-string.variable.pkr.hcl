variable "foo" {
    type = string
    default = "bar"
}

source "null" "test" {
    communicator = "none"
}

build {
    sources = ["null.test"]
}
