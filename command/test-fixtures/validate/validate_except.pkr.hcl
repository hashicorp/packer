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
    inline = [ "echo apple 'apple'" ]
  }

  post-processor "shell-local" {
    name = "pear"
    inline = [ "echo apple 'pear'" ]
  }

  post-processor "shell-local" {
    name = "banana"
  }
}
