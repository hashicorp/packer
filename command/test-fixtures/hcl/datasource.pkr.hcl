data "mock" "content" {
  foo = "chocolate"
}

source "file" "chocolate" {
  content = data.mock.content.foo
  target = "chocolate.txt"
}

build {
  sources = [
    "sources.file.chocolate",
  ]
}
