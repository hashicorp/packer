# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
# Fetch from https://golang.org/dl
TARBALL="https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz"

UNTARPATH="/opt"
GOROOT="${UNTARPATH}/go"
GOPATH="${UNTARPATH}/gopath"

# Install Go
if [ ! -d ${GOROOT} ]; then
  sudo wget --progress=bar:force --output-document - ${TARBALL} |\
    tar xfz - -C ${UNTARPATH}
fi

# Setup the GOPATH
sudo mkdir -p ${GOPATH}
cat <<EOF >/tmp/gopath.sh
export GOROOT="${GOROOT}"
export GOPATH="${GOPATH}"
export PATH="${GOROOT}/bin:${GOPATH}/bin:\$PATH"
EOF
sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh

# Make sure the GOPATH is usable by vagrant
sudo chown -R vagrant:vagrant ${GOROOT}
sudo chown -R vagrant:vagrant ${GOPATH}

# Install some other stuff we need
sudo apt-get update
sudo apt-get install -y curl make git mercurial bzr zip
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
