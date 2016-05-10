# FAT Filesystem Library for Go

This library implements the ability to create, read, and write
FAT filesystems using pure Go.

**WARNING:** While the implementation works (to some degree, see the
limitations section below), I highly recommend you **don't** use this
library, since it has many limitations and is generally a terrible
implementation of FAT. For educational purposes, however, this library
may be interesting.

In this library's current state, it is very good for _reading_ FAT
filesystems, and minimally useful for _creating_ FAT filesystems. See
the features and limitations below.

## Features & Limitations

Features:

* Format a brand new FAT filesystem on a file backed device
* Create files and directories
* Traverse filesystem

Limitations:

This library has several limitations. They're easily able to be overcome,
but because I didn't need them for my use case, I didn't bother:

* Files/directories cannot be deleted or renamed.
* Files never shrink in size.
* Deleted file/directory entries are never reclaimed, so fragmentation
  grows towards infinity. Eventually, your "disk" will become full even
  if you just create and delete a single file.
* There are some serious corruption possibilities in error cases. Cleanup
  is not good.
* Incomplete FAT32 implementation (although FAT12 and FAT16 are complete).

## Usage

Here is some example usage where an existing disk image is read and
a file is created in the root directory:

```go
// Assume this file was created already with a FAT filesystem
f, err := os.OpenFile("FLOPPY.dmg", os.O_RDWR|os.O_CREATE, 0666)
if err != nil {
	panic(err)
}
defer f.Close()

// BlockDevice backed by a file
device, err := fs.NewFileDisk(f)
if err != nil {
	panic(err)
}

filesys, err := fat.New(device)
if err != nil {
	panic(err)
}

rootDir, err := filesys.RootDir()
if err != nil {
	panic(err)
}

subEntry, err := rootDir.AddFile("HELLO_WORLD")
if err != nil {
	panic(err)
}

file, err := subEntry.File()
if err != nil {
	panic(err)
}

_, err = io.WriteString(file, "I am the contents of this file.")
if err != nil {
	panic(err)
}
```

## Thanks

Thanks to the following resources which helped in the creation of this
library:

* [fat32-lib](https://code.google.com/p/fat32-lib/)
* [File Allocation Table on Wikipedia](http://en.wikipedia.org/wiki/File_Allocation_Table)
* Microsoft FAT filesystem specification
