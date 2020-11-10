variable "token" {
  type = string
}

variable "project" {
  type = string
}

source "hyperone" "new-syntax" {
  token = var.token
  project = var.project
  source_image = "debian"
  disk_size = 10
  vm_type = "a1.nano"
  image_name = "packerbats-hcl-{{timestamp}}"
  image_tags = {
      key="value"
  }
}

build {
  sources = [
    "source.hyperone.new-syntax"
  ]

  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get upgrade -y"
    ]
  }
}