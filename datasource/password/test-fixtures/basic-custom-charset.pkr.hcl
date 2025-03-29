source "null" "example" {
  communicator = "none"
}

data "password" "sample" {
  length = 16
  override_special = "pasword"
  special = true
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
