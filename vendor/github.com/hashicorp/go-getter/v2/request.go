package getter

import "net/url"

type Request struct {
	// Src is the source URL to get.
	//
	// Dst is the path to save the downloaded thing as. If Dir is set to
	// true, then this should be a directory. If the directory doesn't exist,
	// it will be created for you.
	//
	// Pwd is the working directory for detection. If this isn't set, some
	// detection may fail. Client will not default pwd to the current
	// working directory for security reasons.
	Src string
	Dst string
	Pwd string

	// Mode is the method of download the client will use. See Mode
	// for documentation.
	Mode Mode

	// Copy, in local file mode if set to true, will copy data instead of using
	// a symlink. If false, attempts to symlink to speed up the operation and
	// to lower the disk space usage. If the symlink fails, may attempt to copy
	// on windows.
	Copy bool

	// Inplace, in local file mode if set to true, do nothing and the returned
	// operation will simply contain the source file path. Inplace has precedence
	// over Copy.
	Inplace bool

	// ProgressListener allows to track file downloads.
	// By default a no op progress listener is used.
	ProgressListener ProgressTracker

	u *url.URL
}
