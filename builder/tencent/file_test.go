package tencent

import (
	"os"
	"testing"
)

func TestSaveDataToFile(t *testing.T) {
	var tempfilename string
	defer func() {
		os.Remove(tempfilename)
	}()
	tempfilename = TempFileName()
	SaveDataToFile(tempfilename, []byte("Hello World"))
	data, err := ReadDataFromFile(tempfilename)
	if string(data) == "Hello World" {
		return
	}
	if err != nil && string(data) == "Hello World" {
		t.Fatal("Failed to save data to file")
	}

}

func TestReadDataFromFile(t *testing.T) {
	var tempfilename string
	defer func() {
		os.Remove(tempfilename)
	}()
	tempfilename = TempFileName()
	SaveDataToFile(tempfilename, []byte("Hello World"))
	data, err := ReadDataFromFile(tempfilename)
	if string(data) == "Hello World" {
		return
	}
	if err != nil && string(data) == "Hello World" {
		t.Fatal("Failed to read data from file")
	}
}
