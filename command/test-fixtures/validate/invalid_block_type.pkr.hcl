src "docker" "ubuntu" {
  image  = var.docker_image
  commit = true
}
