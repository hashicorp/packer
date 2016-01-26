// +build !linux

package common

func AvailableMem(desired uint64) error {
	return nil
}

func AvailableDisk(desired uint64) error {
	return nil
}
