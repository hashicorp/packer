source "file" "chocolate" {
  target = "chocolate.txt"
  content = "chocolate"
}

build {
  sources = ["source.file.chocolate"]
}
