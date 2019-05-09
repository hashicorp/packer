variable "name" {
  default = "packer-www"
  description = "Name of the website in slug format."
}

variable "github_repo" {
  default = "hashicorp/packer"
  description = "GitHub repository of the provider in 'org/name' format."
}

variable "github_branch" {
  default = "stable-website"
  description = "GitHub branch which netlify will continuously deploy."
}

variable "custom_site_domain" {
  default = "packer.io"
  description = "The custom domain to use for the Netlify site."
}