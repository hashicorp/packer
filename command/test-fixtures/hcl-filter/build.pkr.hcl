source "file" "chocolate" {
  content = "chocolate"
  target  = "chocolate.txt"
  metadata {
    tags   = ["prod", "x86"]
    labels = { region = "us-east" }
  }
}

source "file" "vanilla" {
  content = "vanilla"
  target  = "vanilla.txt"
  metadata {
    tags   = ["dev", "x86"]
    labels = { region = "us-west" }
  }
}

source "file" "cherry" {
  content = "cherry"
  target  = "cherry.txt"
  metadata {
    tags   = ["prod", "arm64"]
    labels = { region = "eu-west" }
  }
}

build {
  name = "my_build"
  metadata {
    tags = ["nightly"]
  }
  sources = [
    "file.chocolate",
    "file.vanilla",
    "file.cherry",
  ]
}
