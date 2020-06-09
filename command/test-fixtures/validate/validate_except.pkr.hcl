source "file" "chocolate" {
  content = "chocolate"
  target = "chocolate.txt"
}
source "file" "vanilla" {
  content = "vanilla"
}

build {
  sources = [
    "source.file.chocolate",
    "source.file.vanilla"
  ]

  post-processor "shell-local" {
    name = "apple"
    inline = [ "echo apple 'hello'" ]
  }

  post-processor "shell-local" {
    name = "pear"
  }
}
