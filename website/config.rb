#-------------------------------------------------------------------------
# Configure Middleman
#-------------------------------------------------------------------------

set :base_url, "https://www.packer.io/"

activate :hashicorp do |h|
  h.version      = '0.7.1'
  h.bintray_repo = 'mitchellh/packer'
  h.bintray_user = 'mitchellh'
  h.bintray_key  = ENV['BINTRAY_API_KEY']
end
