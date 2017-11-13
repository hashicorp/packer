package getter

import (
	"log"
	"net/url"
	"os"
	"time"
)

// FileGetter is a Getter implementation that will download a module from
// a file scheme.
type FileGetter struct {
	// Copy, if set to true, will copy data instead of using a symlink
	Copy bool

	// Used for calculating percent progress
	totalSize       int64
	PercentComplete int
	Done            chan int64
}

func (g *FileGetter) ClientMode(u *url.URL) (ClientMode, error) {
	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	// Check if the source is a directory.
	if fi.IsDir() {
		return ClientModeDir, nil
	}

	return ClientModeFile, nil
}

func (g *FileGetter) CalcDownloadPercent(dst string) {
	// stat file every n seconds to figure out the download progress
	var stop bool = false
	dstfile, err := os.Open(dst)
	defer dstfile.Close()

	if err != nil {
		log.Printf("couldn't open file for reading: %s", err)
		return
	}
	for {
		select {
		case <-g.Done:
			stop = true
		default:
			g.PercentComplete = CalcPercent(dstfile, g.totalSize)
		}

		if stop {
			break
		}
		// repeat check once per second
		time.Sleep(time.Second)
	}
}
