#!/bin/sh -eux

# set a default HOME_DIR environment variable if not set
HOME_DIR="${HOME_DIR:-/home/vagrant}";

case "$PACKER_BUILDER_TYPE" in
parallels-iso|parallels-pvm)
    mkdir -p /tmp/parallels;
    mount -o loop $HOME_DIR/prl-tools-lin.iso /tmp/parallels;
    VER="`cat /tmp/parallels/version`";

    echo "Parallels Tools Version: $VER";

    /tmp/parallels/install --install-unattended-with-deps \
      || (code="$?"; \
          echo "Parallels tools installation exited $code, attempting" \
          "to output /var/log/parallels-tools-install.log"; \
          cat /var/log/parallels-tools-install.log; \
          exit $code);
    umount /tmp/parallels;
    rm -rf /tmp/parallels;
    rm -f $HOME_DIR/*.iso;

    # Parallels Tools for Linux includes native auto-mount script,
    # which causes losing some of Vagrant-relative shared folders.
    # So, we should disable this behavior.
    # https://github.com/Parallels/vagrant-parallels/issues/325#issuecomment-418727113
    auto_mount_script='/usr/bin/prlfsmountd'
    if [ -f "${auto_mount_script}" ]; then
        echo -e '#!/bin/sh\n'\
        '# Shared folders auto-mount is disabled by Vagrant ' \
        > "${auto_mount_script}"
    fi
    ;;
esac
