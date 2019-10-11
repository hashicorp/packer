package cvm

import "testing"

func TestTencentCloudImageConfig_Prepare(t *testing.T) {
	cf := &TencentCloudImageConfig{
		ImageName: "foo",
	}

	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %v", err)
	}

	cf.ImageName = "foo.:"
	if err := cf.Prepare(nil); err != nil {
		t.Fatal("shouldn't have error")
	}

	cf.ImageName = "foo"
	cf.ImageCopyRegions = []string{"ap-guangzhou", "ap-hongkong"}
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %v", err)
	}

	cf.ImageCopyRegions = []string{"unknown"}
	if err := cf.Prepare(nil); err == nil {
		t.Fatal("should have err")
	}

	cf.SkipValidation = true
	if err := cf.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err:%v", err)
	}

}
