
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

build {
  dynamic "source" {
    for_each = var.images
    labels   = ["file.base"]
    content {
      name         = source.key
      target       = source.value.image
      content      = join("\n", formatlist("layers/%s/files", var.images[source.key].layers))
    }
  }
}
