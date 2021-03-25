
source "file" "base" {
}

variables {
  images = {
    dummy = {
      image      = "dummy"
      layers     = ["base/main"]
    }
    postgres = {
      image      = "postgres/13"
      layers     = ["base/main", "base/init", "postgres"]
    }
  }
}

locals {
  files = {
    foo = {
      destination = "fooo"
    }
    bar = {
      destination = "baar"
    }
  }
}

build {
  dynamic "source" {
    for_each = var.images
    labels   = ["file.base"]
    content {
      name         = source.key
      target       = "${path.root}/${source.value.image}"
      content      = join("\n", formatlist("layers/%s/files", var.images[source.key].layers))
    }
  }

  dynamic "provisioner" {
    for_each = local.files
    labels   = ["shell-local"]
    content {
      inline = ["echo 1 > ${path.root}/${var.images[source.name].image}-${provisioner.value.destination}"]
    }
  }
}
