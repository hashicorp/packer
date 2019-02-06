### SSH key pair automation

The VirtualBox builders can inject the current SSH key pair's public key into
the template using the following variables:

-   `SSHPublicKey` (*VirtualBox builders only*) - The SSH public key as a line
    in OpenSSH authorized_keys format.
-   `EncodedSSHPublicKey` (*VirtualBox builders only*) - The same as
    `SSHPublicKey`, except it is URL encoded for usage in places
    like the kernel command line.

When a private key is provided using `ssh_private_key_file`, the key's
corresponding public key can be accessed using the above variables.

If `ssh_password` and `ssh_private_key_file` are not specified, Packer will
automatically generate en ephemeral key pair. The key pair's public key can
be accessed using the template variables.

For example, the public key can be provided in the boot command:
```json
{
    "type": "virtualbox-iso",
    "boot_command": [
      "<up><wait><tab> text ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/ks.cfg PACKER_USER={{ user `username` }} PACKER_AUTHORIZED_KEY={{ .EncodedSSHPublicKey }}<enter>"
    ]
}
```

The kickstart can then leverage those fields from the kernel command line:
```
%post

# Newly created users need the file/folder framework for SSH key authentication.
umask 0077
mkdir /etc/skel/.ssh
touch /etc/skel/.ssh/authorized_keys

# Loop over the command line. Set interesting variables.
for x in $(cat /proc/cmdline)
do
  case $x in
    PACKER_USER=*)
      PACKER_USER="${x#*=}"
      ;;
    PACKER_AUTHORIZED_KEY=*)
      encoded="${x#*=}"
      # URL decode $encoded into $PACKER_AUTHORIZED_KEY
      printf -v PACKER_AUTHORIZED_KEY '%b' "${encoded//%/\\x}"
      ;;
  esac
done

# Create/configure packer user, if any.
if [ -n "$PACKER_USER" ]
then
  useradd $PACKER_USER
  echo "%$PACKER_USER ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers.d/$PACKER_USER
  [ -n "$PACKER_AUTHORIZED_KEY" ] && echo $PACKER_AUTHORIZED_KEY >> $(eval echo ~"$PACKER_USER")/.ssh/authorized_keys
fi

%end
```
