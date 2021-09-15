
variable "bkt" {
    default = "pkr-acctest-temp-1"
}

data "hcp-packer-iteration" "acc" {
    bucket_name = var.bkt
    channel     = "acc"
}

data "hcp-packer-image" "acc-production" {
    bucket_name    = var.bkt
    iteration_id   = data.hcp-packer-iteration.acc.id
    cloud_provider = "aws"
    region         = "us-west-1"
}

source "null" "example" {
    communicator = "none"
}

build {
	sources = [
		"source.null.example"
	]
	provisioner "shell-local" {
		inline = ["echo the artifact id is: '${data.hcp-packer-image.acc-production.id}', yup yup"]
	}
}
