#!/bin/sh -eux

# set a default HOME_DIR environment variable if not set
HOME_DIR="${HOME_DIR:-/home/vagrant}";

case "$PACKER_BUILDER_TYPE" in
vmware-iso|vmware-vmx)

    # make sure we have /sbin in our path. RHEL systems lack this
    PATH=/sbin:$PATH
    export PATH

    mkdir -p /tmp/vmware;
    mkdir -p /tmp/vmware-archive;
    mount -o loop $HOME_DIR/linux.iso /tmp/vmware;

    TOOLS_PATH="`ls /tmp/vmware/VMwareTools-*.tar.gz`";
    VER="`echo "${TOOLS_PATH}" | cut -f2 -d'-'`";
    MAJ_VER="`echo ${VER} | cut -d '.' -f 1`";

    echo "VMware Tools Version: $VER";

    tar xzf ${TOOLS_PATH} -C /tmp/vmware-archive;
    if [ "${MAJ_VER}" -lt "10" ]; then
        /tmp/vmware-archive/vmware-tools-distrib/vmware-install.pl --default;
    else
        /tmp/vmware-archive/vmware-tools-distrib/vmware-install.pl --force-install;
    fi
    umount /tmp/vmware;
    rm -rf  /tmp/vmware;
    rm -rf  /tmp/vmware-archive;
    rm -f $HOME_DIR/*.iso;
    ;;
esac
