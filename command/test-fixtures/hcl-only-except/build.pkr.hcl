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
  source "file.cherry" {

  }
}

build {
  name = "my_build"
  sources = [
    "file.chocolate",
    "file.vanilla",
  ]
}
