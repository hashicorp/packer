data "file" "test" {
  contents = "Hello there!"
  destination = "./out_dir/subdir/"
}

source "null" "test" {
  communicator = "none"
}

build {
  sources = ["null.test"]
}
