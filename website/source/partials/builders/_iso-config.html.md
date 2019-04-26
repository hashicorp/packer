## ISO Configuration Reference

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


\~&gt; On windows - when referencing a local iso - if packer is running without
symlinking rights, the iso will be copied to the cache folder. Read [Symlinks
in Windows 10
!](https://blogs.windows.com/buildingapps/2016/12/02/symlinks-windows-10/) for
more info.

### Required:

-   `iso_checksum` (string) - The checksum for the ISO file or virtual hard
    drive file. The algorithm to use when computing the checksum can be
    optionally specified with `iso_checksum_type`. When `iso_checksum_type` is
    not set packer will guess the checksumming type based on `iso_checksum`
    length. `iso_checksum` can be also be a file or an URL, in which case
    `iso_checksum_type` must be set to `file`; the go-getter will download it
    and use the first hash found.

-   `iso_url` (string) - A URL to the ISO containing the installation image or
    virtual hard drive (VHD or VHDX) file to clone.

### Optional:

-   `iso_checksum_type` (string) - The algorithm to be used when computing the
    checksum of the file specified in `iso_checksum`. Currently, valid values
    are "", "none", "md5", "sha1", "sha256", "sha512" or "file". Since the
    validity of ISO and virtual disk files are typically crucial to a
    successful build, Packer performs a check of any supplied media by default.
    While setting "none" will cause Packer to skip this check, corruption of
    large files such as ISOs and virtual hard drives can occur from time to
    time. As such, skipping this check is not recommended. `iso_checksum_type`
    must be set to `file` when `iso_checksum` is an url.

-   `iso_checksum_url` (string) - A URL to a checksum file containing a
    checksum for the ISO file. At least one of `iso_checksum` and
    `iso_checksum_url` must be defined. `iso_checksum_url` will be ignored if
    `iso_checksum` is non empty.

-   `iso_target_extension` (string) - The extension of the iso file after
    download. This defaults to `iso`.

-   `iso_target_path` (string) - The path where the iso should be saved after
    download. By default will go in the packer cache, with a hash of the
    original filename and checksum as its name.

-   `iso_urls` (array of strings) - Multiple URLs for the ISO to download.
    Packer will try these in order. If anything goes wrong attempting to
    download or while downloading a single URL, it will move on to the next.
    All URLs must point to the same file (same checksum). By default this is
    empty and `iso_url` is used. Only one of `iso_url` or `iso_urls` can be
    specified.

### Example ISO configurations

go-getter can guess the checksum type based on `iso_checksum` len.

``` json
{ 
  "iso_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2", 
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```

``` json
{ 
  "iso_checksum_type": "file",
  "iso_checksum": "ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum",
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```

``` json
{ 
  "iso_checksum_url": "./shasums.txt", 
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```

``` json
{ 
  "iso_checksum_type": "sha256",
  "iso_checksum_url": "./shasums.txt",
  "iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"
}
```
