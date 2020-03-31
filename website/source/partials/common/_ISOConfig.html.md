<!-- Code generated from the comments of the ISOConfig struct in common/iso_config.go; DO NOT EDIT MANUALLY -->
By default, Packer will symlink, download or copy image files to the Packer
cache into a "`hash($iso_url+$iso_checksum).$iso_target_extension`" file.
Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
file mode in order to perform a download.

go-getter supports the following protocols:

* Local files
* Git
* Mercurial
* HTTP
* Amazon S3

Examples:
go-getter can guess the checksum type based on `iso_checksum` len.

```json
{
  "iso_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```

```json
{
  "iso_checksum_type": "file",
  "iso_checksum": "ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum",
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```

```json
{
  "iso_checksum_url": "./shasums.txt",
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```

```json
{
  "iso_checksum_type": "sha256",
  "iso_checksum_url": "./shasums.txt",
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```
