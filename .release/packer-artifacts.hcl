# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

schema = 1
artifacts {
  zip = [
    "packer_${version}_darwin_amd64.zip",
    "packer_${version}_darwin_arm64.zip",
    "packer_${version}_freebsd_386.zip",
    "packer_${version}_freebsd_amd64.zip",
    "packer_${version}_freebsd_arm.zip",
    "packer_${version}_linux_386.zip",
    "packer_${version}_linux_amd64.zip",
    "packer_${version}_linux_arm.zip",
    "packer_${version}_linux_arm64.zip",
    "packer_${version}_linux_ppc64le.zip",
    "packer_${version}_netbsd_386.zip",
    "packer_${version}_netbsd_amd64.zip",
    "packer_${version}_netbsd_arm.zip",
    "packer_${version}_openbsd_386.zip",
    "packer_${version}_openbsd_amd64.zip",
    "packer_${version}_openbsd_arm.zip",
    "packer_${version}_solaris_amd64.zip",
    "packer_${version}_windows_386.zip",
    "packer_${version}_windows_amd64.zip",
  ]
  rpm = [
    "packer-${version_linux}-1.aarch64.rpm",
    "packer-${version_linux}-1.armv7hl.rpm",
    "packer-${version_linux}-1.i386.rpm",
    "packer-${version_linux}-1.ppc64le.rpm",
    "packer-${version_linux}-1.x86_64.rpm",
  ]
  deb = [
    "packer_${version_linux}-1_amd64.deb",
    "packer_${version_linux}-1_arm64.deb",
    "packer_${version_linux}-1_armhf.deb",
    "packer_${version_linux}-1_i386.deb",
    "packer_${version_linux}-1_ppc64el.deb",
  ]
  container = [
    "packer_release-full_linux_386_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-full_linux_386_${version}_${commit_sha}.docker.tar",
    "packer_release-full_linux_amd64_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-full_linux_amd64_${version}_${commit_sha}.docker.tar",
    "packer_release-full_linux_arm64_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-full_linux_arm64_${version}_${commit_sha}.docker.tar",
    "packer_release-full_linux_arm_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-full_linux_arm_${version}_${commit_sha}.docker.tar",
    "packer_release-light_linux_386_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-light_linux_386_${version}_${commit_sha}.docker.tar",
    "packer_release-light_linux_amd64_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-light_linux_amd64_${version}_${commit_sha}.docker.tar",
    "packer_release-light_linux_arm64_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-light_linux_arm64_${version}_${commit_sha}.docker.tar",
    "packer_release-light_linux_arm_${version}_${commit_sha}.docker.dev.tar",
    "packer_release-light_linux_arm_${version}_${commit_sha}.docker.tar",
  ]
}
