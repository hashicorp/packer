variable "target" {
 type    = string
 default = %q
}

source "file" "test" {
  content = "Lorem ipsum dolor sit amet"
  target  = var.target
}

build {
    hcp_packer_registry {
     bucket_name = "simple-deprecated"
    }
  sources = ["source.file.test"]
}
