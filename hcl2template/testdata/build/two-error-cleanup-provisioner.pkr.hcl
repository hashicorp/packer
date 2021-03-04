source "virtualbox-iso" "ubuntu-1204" {
}

// starts resources to provision them.
build {
  sources = [
    "source.virtualbox-iso.ubuntu-1204"
  ]

  error-cleanup-provisioner "shell-local" {
  }

  error-cleanup-provisioner "file" {
  }
}
