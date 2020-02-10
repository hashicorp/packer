#!/bin/sh
set -eux

# Based on
# https://gist.github.com/agentsim/00cc38c693e7d0e1b36a2080870d955b#gistcomment-2304505

mkdir -p out

hdiutil create -o out/HighSierra.cdr -size 5530m -layout SPUD -fs HFS+J
hdiutil attach out/HighSierra.cdr.dmg -noverify -mountpoint /Volumes/install_build
sudo /Applications/Install\ macOS\ High\ Sierra.app/Contents/Resources/createinstallmedia --volume /Volumes/install_build --nointeraction
hdiutil detach /Volumes/Install\ macOS\ High\ Sierra
hdiutil convert out/HighSierra.cdr.dmg -format UDTO -o out/HighSierra.iso
mv out/HighSierra.iso.cdr out/HighSierra.iso
rm out/HighSierra.cdr.dmg
