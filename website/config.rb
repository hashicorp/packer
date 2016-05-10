set :base_url, "https://www.packer.io/"

activate :hashicorp do |h|
  h.name        = "packer"
  h.version     = "0.10.1"
  h.github_slug = "mitchellh/packer"
end
