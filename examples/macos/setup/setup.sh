#!/bin/sh
set -eux

# Format partition
diskutil eraseDisk JHFS+ Disk disk0

# Packages are installed in reversed order - why?
"/Volumes/Image Volume/Install macOS High Sierra.app/Contents/Resources/startosinstall" \
  --volume /Volumes/Disk \
  --converttoapfs no \
  --agreetolicense \
  --installpackage "/Volumes/setup/postinstall.pkg" \
  --installpackage "/Volumes/VMware Tools/Install VMware Tools.app/Contents/Resources/VMware Tools.pkg"
