#!/bin/sh

set -ex

apk add libressl
apk add open-vm-tools
rc-update add open-vm-tools
/etc/init.d/open-vm-tools start

cat >/usr/local/bin/shutdown <<EOF
#!/bin/sh
poweroff
EOF
chmod +x /usr/local/bin/shutdown

sed -i "/#PermitRootLogin/c\PermitRootLogin yes" /etc/ssh/sshd_config
mkdir ~/.ssh
wget https://raw.githubusercontent.com/jetbrains-infra/packer-builder-vsphere/master/test/test-key.pub -O ~/.ssh/authorized_keys
/etc/init.d/sshd restart
