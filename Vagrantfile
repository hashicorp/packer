Vagrant.configure(2) do |config|
  config.vm.box = "cbednarski/ubuntu-1404-dev"
  config.vm.provider "virtualbox" do |vb|
    vb.memory = "4096"
    vb.cpus = "4"
  end
  config.vm.provider "vmware_desktop" do |v|
    v.vmx["memsize"] = "4096"
    v.vmx["numvcpus"] = "4"
    v.vmx["cpuid.coresPerSocket"] = "1"
  end

  config.vm.provision "shell", inline: <<-SHELL
    sudo ifs aptupdate
    sudo ifs install golang
    sudo ifs install docker

    grep GOPATH ~/.bashrc > /dev/null || echo '
export GOPATH=\$HOME/go
export GOBIN=\$GOPATH/bin
export PATH=\$PATH:\$GOBIN
cd \$GOPATH/src/github.com/mitchellh/packer
' >> ~/.bashrc
  SHELL

  config.vm.provision "shell", inline: <<-SHELL
    go get github.com/mitchellh/packer
    cd $GOPATH/src/github.com/mitchellh/packer
    make updatedeps
    make dev
    packer version
  SHELL
end