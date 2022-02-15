
data "null" "secret" {
  input = "s3cr3t"
}

locals {
  secret = data.null.secret.output
}

source "file" "foo" {
  content = "foo"
  target = "foo.txt"
}

build {
  sources = ["file.foo"]
  provisioner "shell-local" {
    # original bug in :
    # environment_vars = ["MY_SECRET=${local.secret}"]
    env = {
      "MY_SECRET":"${local.secret}",
    }
    inline           = ["echo yo, my secret is $MY_SECRET"]
  }
}
