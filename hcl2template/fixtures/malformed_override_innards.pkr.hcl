provisioner "shell-local" {
  inline = ["echo 'hi'"]
  override = {
    test = "hello"
  }
}
