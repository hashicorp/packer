#-------------------------------------------------------------------------
# Configure Middleman
#-------------------------------------------------------------------------

set :base_url, "https://www.packer.io/"

activate :hashicorp do |h|
  h.version         = ENV["PACKER_VERSION"]
  h.bintray_enabled = ENV["BINTRAY_ENABLED"]
  h.bintray_repo    = "mitchellh/packer"
  h.bintray_user    = "mitchellh"
  h.bintray_key     = ENV["BINTRAY_API_KEY"]
end
