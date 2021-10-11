source "null" "example" {
  communicator = "none"
}

build {
  name    = "example"
  sources = ["source.null.example"]

  post-processor "shell-local" {
    inline = ["echo 2 > ${build.name}.2.txt"]
  }
}
