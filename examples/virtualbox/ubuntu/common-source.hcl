
source "virtualbox-iso" "base-ubuntu" {
    boot_command = [
        "<esc><wait>",
        "<esc><wait>",
        "<enter><wait>",
        "/install/vmlinuz<wait>",
        " auto<wait>",
        " console-setup/ask_detect=false<wait>",
        " console-setup/layoutcode=us<wait>",
        " console-setup/modelcode=pc105<wait>",
        " debconf/frontend=noninteractive<wait>",
        " debian-installer=en_US.UTF-8<wait>",
        " fb=false<wait>",
        " initrd=/install/initrd.gz<wait>",
        " kbd-chooser/method=us<wait>",
        " keyboard-configuration/layout=USA<wait>",
        " keyboard-configuration/variant=USA<wait>",
        " locale=en_US.UTF-8<wait>",
        " netcfg/get_domain=vm<wait>",
        " netcfg/get_hostname=vagrant<wait>",
        " grub-installer/bootdev=/dev/vda<wait>",
        " noapic<wait>",
        " preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/{{user `preseed_path`}}<wait>",
        " -- <wait>",
        "<enter><wait>"
    ]
    boot_wait = "10s"
    guest_os_type = "Ubuntu_64"
    http_directory = "etc/"
    shutdown_command = "echo 'vagrant' | sudo -S shutdown -P now"
    ssh_password = "vagrant"
    ssh_port = 22
    ssh_username = "vagrant"
    ssh_wait_timeout = "10000s"
}
