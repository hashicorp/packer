package driver

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func TestDatastoreAcc(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	d := newTestDriver(t)
	ds, err := d.FindDatastore("datastore1", "")
	if err != nil {
		t.Fatalf("Cannot find the default datastore '%v': %v", "datastore1", err)
	}
	info, err := ds.Info("name")
	if err != nil {
		t.Fatalf("Cannot read datastore properties: %v", err)
	}
	if info.Name != "datastore1" {
		t.Errorf("Wrong datastore. expected: 'datastore1', got: '%v'", info.Name)
	}
}

func TestFileUpload(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	dsName := "datastore1"
	hostName := "esxi-1.vsphere65.test"

	fileName := fmt.Sprintf("test-%v", time.Now().Unix())
	tmpFile, err := ioutil.TempFile("", fileName)
	if err != nil {
		t.Fatalf("Error creating temp file")
	}
	err = tmpFile.Close()
	if err != nil {
		t.Fatalf("Error creating temp file")
	}

	d := newTestDriver(t)
	ds, err := d.FindDatastore(dsName, hostName)
	if err != nil {
		t.Fatalf("Cannot find datastore '%v': %v", dsName, err)
	}

	err = ds.UploadFile(tmpFile.Name(), fileName, hostName, true)
	if err != nil {
		t.Fatalf("Cannot upload file: %v", err)
	}

	if ds.FileExists(fileName) != true {
		t.Fatalf("Cannot find file")
	}

	err = ds.Delete(fileName)
	if err != nil {
		t.Fatalf("Cannot delete file: %v", err)
	}
}

func TestFileUploadDRS(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	dsName := "datastore3"
	hostName := ""

	fileName := fmt.Sprintf("test-%v", time.Now().Unix())
	tmpFile, err := ioutil.TempFile("", fileName)
	if err != nil {
		t.Fatalf("Error creating temp file")
	}
	err = tmpFile.Close()
	if err != nil {
		t.Fatalf("Error creating temp file")
	}

	d := newTestDriver(t)
	ds, err := d.FindDatastore(dsName, hostName)
	if err != nil {
		t.Fatalf("Cannot find datastore '%v': %v", dsName, err)
	}

	err = ds.UploadFile(tmpFile.Name(), fileName, hostName, false)
	if err != nil {
		t.Fatalf("Cannot upload file: %v", err)
	}

	if ds.FileExists(fileName) != true {
		t.Fatalf("Cannot find file")
	}

	err = ds.Delete(fileName)
	if err != nil {
		t.Fatalf("Cannot delete file: %v", err)
	}
}
