package getter

import (
	"fmt"
	"os"
)

func CalcPercent(dstfile *os.File, totalSize int64) int {
	fi, err := dstfile.Stat()
	if err != nil {
		fmt.Printf("Error stating file: %s", err)
		return 101
	}
	size := fi.Size()

	// catch edge case that would break our percentage calc
	if size == 0 {
		size = 1
	}
	return int(float64(size) / float64(totalSize) * 100)
}
