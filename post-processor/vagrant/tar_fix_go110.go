// +build go1.10

package vagrant

import "archive/tar"

func setHeaderFormat(header *tar.Header) {
	// We have to set the Format explicitly because of a bug in
	// libarchive. This affects eg. the tar in macOS listing huge
	// files with zero byte length.
	header.Format = tar.FormatGNU
}
