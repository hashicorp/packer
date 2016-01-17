# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
TARBALL="https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz"

# Install Go
sudo wget --progress=bar:force --output-document - ${TARBALL} |\
  tar xfz - -C /opt

# Setup the GOPATH
sudo mkdir -p /opt/gopath
cat <<EOF >/tmp/gopath.sh
export GOROOT="/opt/go"
export GOPATH="/opt/gopath"
export PATH="/opt/go/bin:/opt/gopath/bin:\$PATH"
EOF
sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh

# Make sure the gopath is usable by vagrant
sudo chown -R vagrant:vagrant /opt/go
sudo chown -R vagrant:vagrant /opt/gopath

# Install some other stuff we need
sudo apt-get update
sudo apt-get install -y curl git mercurial bzr zip
SCRIPT

Vagrant.configure(2) do |config|
  config.vm.box = "bento/ubuntu-14.04"

  config.vm.provision "shell", inline: $script

  config.vm.synced_folder ".", "/vagrant", disabled: true

  ["vmware_fusion", "vmware_workstation"].each do |p|
    config.vm.provider "p" do |v|
      v.vmx["memsize"] = "2048"
      v.vmx["numvcpus"] = "2"
      v.vmx["cpuid.coresPerSocket"] = "1"
    end
  end
end
