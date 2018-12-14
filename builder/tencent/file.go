package tencent

import (
	"io/ioutil"
)

// SaveDataToFile is a simple way of writing to a given filename.
// Returns true if successful
// If the filename given points to an existing file, its contents are overwritten.
func SaveDataToFile(filename string, data []byte) (bool, error) {
	err := ioutil.WriteFile(filename, data, 0644)
	return err == nil, err
}

// ReadDataFromFile reads the contents of the given filename into a byte array and returns it.
func ReadDataFromFile(filename string) ([]byte, error) {
	result, err := ioutil.ReadFile(filename)
	return result, err
}
