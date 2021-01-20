packer {
  required_version = ">= v1.0.0"
}

data "mock" "content" {
  foo = "content"
}

source "file" "chocolate" {
  target = "chocolate.txt"
  content = data.mock.content.foo
}

build {
  sources = ["source.file.chocolate"]
}
