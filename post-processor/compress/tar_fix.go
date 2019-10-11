// +build !go1.10

package compress

import "archive/tar"

func setHeaderFormat(header *tar.Header) {
	// no-op
}
