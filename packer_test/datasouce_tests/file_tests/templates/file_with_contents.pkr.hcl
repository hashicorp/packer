data "file" "test" {
  contents = "Hello there!"
}

source "null" "test" {
  communicator = "none"
}

build {
  sources = ["null.test"]
}
