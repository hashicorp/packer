source "null" "example" {
  communicator = "none"
}

data "http" "basic" {
  url = "https://www.packer.io/"
  method = "NONEEXISTING"
}

locals {
  url = "${data.http.basic.url}"
  body = "${data.http.basic.body}" != ""
}

build {
  name = "mybuild"
  sources = [
    "source.null.example"
  ]
  provisioner "shell-local" {
    inline = [
      "echo url is ${local.url}",
      "echo body is ${local.body}"
    ]
  }
}
