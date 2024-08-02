# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

schema = 1
artifacts {
  zip = [
    "levant_${version}_darwin_amd64.zip",
    "levant_${version}_darwin_arm64.zip",
    "levant_${version}_freebsd_amd64.zip",
    "levant_${version}_freebsd_arm64.zip",
    "levant_${version}_linux_amd64.zip",
    "levant_${version}_linux_arm64.zip",
    "levant_${version}_windows_amd64.zip",
    "levant_${version}_windows_arm64.zip",
  ]
  rpm = [
    "levant-${version_linux}-1.aarch64.rpm",
    "levant-${version_linux}-1.x86_64.rpm",
  ]
  deb = [
    "levant_${version_linux}-1_amd64.deb",
    "levant_${version_linux}-1_arm64.deb",
  ]
  container = [
    "levant_release_linux_amd64_${version}_${commit_sha}.docker.dev.tar",
    "levant_release_linux_amd64_${version}_${commit_sha}.docker.tar",
    "levant_release_linux_arm64_${version}_${commit_sha}.docker.dev.tar",
    "levant_release_linux_arm64_${version}_${commit_sha}.docker.tar",
  ]
}
