source "null" "test" {
  communicator = "none"
}

build {
  sources = ["null.test"]

  post-processor "manifest" {}
}
