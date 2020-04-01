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
    "file.chocolate",
    "file.vanilla",
    "file.cherry",
  ]
}
