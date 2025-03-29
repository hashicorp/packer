source "null" "example" {
  communicator = "none"
}

data "password" "sample" {
  length = 16
}

build {
  name = "mybuild"
  sources = [
    "source.null.example"
  ]
  provisioner "shell-local" {
    inline = [
      "echo password: '${data.password.sample.result}'",
      "echo generated hash is '${data.password.sample.bcrypt_hash}'"
    ]
  }
}
