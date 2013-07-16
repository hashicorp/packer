require "net/http"

raise "PACKER_VERSION must be set." if !ENV["PACKER_VERSION"]

#-------------------------------------------------------------------------
# Download the list of Packer downloads
#-------------------------------------------------------------------------

$packer_files = {}
$packer_os = []

if !ENV["PACKER_DISABLE_DOWNLOAD_FETCH"]
  raise "BINTRAY_API_KEY must be set." if !ENV["BINTRAY_API_KEY"]
  http = Net::HTTP.new("dl.bintray.com", 80)
  req = Net::HTTP::Get.new("/mitchellh/packer")
  req.basic_auth "mitchellh", ENV["BINTRAY_API_KEY"]
  response = http.request(req)

  response.body.split("\n").each do |line|
    next if line !~ /\/mitchellh\/packer\/(#{Regexp.quote(ENV["PACKER_VERSION"])}.+?)'/
      p $1.to_s
    filename = $1.to_s
    os = filename.split("_")[1]
    next if os == "SHA256SUMS"

    $packer_files[os] ||= []
    $packer_files[os] << filename
  end

  $packer_os = ["darwin", "linux", "windows"] & $packer_files.keys
  $packer_os += $packer_files.keys
  $packer_os.uniq!

  $packer_files.each do |key, value|
    value.sort!
  end
end

#-------------------------------------------------------------------------
# Configure Middleman
#-------------------------------------------------------------------------

set :css_dir, 'stylesheets'
set :js_dir, 'javascripts'
set :images_dir, 'images'

# Use the RedCarpet Markdown engine
set :markdown_engine, :redcarpet
set :markdown, :fenced_code_blocks => true

# Build-specific configuration
configure :build do
  activate :asset_hash
  activate :minify_css
  activate :minify_html
  activate :minify_javascript
end

#-------------------------------------------------------------------------
# Helpers
#-------------------------------------------------------------------------
helpers do
  def download_arch(file)
    parts = file.split("_")
    return "" if parts.length != 3
    parts[2].split(".")[0]
  end

  def download_os_human(os)
    if os == "darwin"
      return "Mac OS X"
    elsif os == "freebsd"
      return "FreeBSD"
    elsif os == "openbsd"
      return "OpenBSD"
    elsif os == "Linux"
      return "Linux"
    elsif os == "windows"
      return "Windows"
    else
      return os
    end
  end

  def download_url(file)
    "https://dl.bintray.com/mitchellh/packer/#{file}?direct"
  end

  def latest_version
    ENV["PACKER_VERSION"]
  end
end
