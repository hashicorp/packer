source "file" "chocolate" {
  content = "chocolate"
  target = "chocolate.txt"
}

source "file" "vanilla" {
  content = "vanilla"
  target = "vanilla.txt"
}

source "file" "cherry" {
  content = "cherry"
  target = "cherry.txt"
}


build {
  sources = [
    "sources.file.chocolate",
    "sources.file.vanilla",
    "sources.file.cherry",
  ]

  post-processor "shell-local" {
    name = "apple"
    inline = [ "echo apple > apple.txt" ]
  }

  post-processor "shell-local" {
    name = "peach"
    inline = [ "echo apple > peach.txt" ]
  }

  post-processor "shell-local" {
    name = "pear"
    inline = [ "echo apple > pear.txt" ]
  }

  post-processor "shell-local" {
    name = "banana"
    inline = [ "echo apple > banana.txt" ]
  }

  post-processor "shell-local" {
    only = ["file.vanilla"]
    name = "tomato"
    inline = [ "echo apple > tomato.txt" ]
  }

  post-processor "shell-local" {
    only = ["file.chocolate"]
    inline = [ "echo apple > unnamed.txt" ]
  }
}
