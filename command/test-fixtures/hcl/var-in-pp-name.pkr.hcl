source "null" "example1" {
  communicator = "none"
}

source "null" "example2" {
  communicator = "none"
}

locals {
  except_example2 = "null.example2"
  true   = true
}

variable "only_example2" {
  default = "null.example2"
}

variable "foo" {
  default = "bar"
}

build {
  sources = ["source.null.example1", "source.null.example2"]
  post-processor "shell-local" {
    keep_input_artifact = local.true
    except              = [local.except_example2]
    inline              = ["echo 1 > ${source.name}.1.txt"]
  }

  post-processor "shell-local" {
    name   = var.foo
    only   = [var.only_example2]
    inline = ["echo 2 > ${source.name}.2.txt"]
  }
}
