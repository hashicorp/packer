packer {
  required_plugins {
    tester = {
      source = "github.com/hashicorp/tester",
      version = ">= 1.0.0"
    }
    pbtester = {
      source = "github.com/hashicorp/pbtester",
      version = ">= 1.0.0"
    }
  }
}

source "tester-dynamic" "test" {}
source "pbtester-dynamic" "test" {}

build {
  sources = [
    "tester-dynamic.test",
    "pbtester-dynamic.test"
  ]
}
