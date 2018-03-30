// +build go1.10

package compress

import "archive/tar"

func setHeaderFormat(header *tar.Header) {
	// We have to set the Format explicitly for the googlecompute-import
	// post-processor. Google Cloud only allows importing GNU tar format.
	header.Format = tar.FormatGNU
}
