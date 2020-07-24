

build {
  name = "aws_example_builder"
  description = <<EOF
The builder of clouds !!

Use it at will.
EOF

  sources = [
      "source.amazon-ebs.example-1",

      // this one is not defined but we don't want to error there, we just
      // would like to show what sources are being referenced.
      "source.amazon-ebs.example-2", 
  ]

  provisioner "shell" {
    files = [
      "bins/install-this.sh",
      "bins/install-that.sh",
      "bins/conf-this.sh",
    ]
  }

  post-processor "manifest" {
  }

  post-processor "shell-local" {
  }

  post-processors {
    post-processor "manifest" {
    }

    post-processor "shell-local" {
    }
  }
}