// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

import (
	"os"
	"testing"
)

func TestNewConfigMissingFile(t *testing.T) {
	// WHEN
	_, err := LoadConfigsFromFile("some/invalid/path")

	// THEN

	if err == nil {
		t.Error("Expected missing file error")
	}
}

func TestNewConfigDefaultOnly(t *testing.T) {
	// GIVEN

	// Get DEFAULT config
	cfg, keyFile, err := BaseTestConfig()
	defer os.Remove(keyFile.Name())

	// Write test config to file
	f, err := WriteTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	// WHEN

	// Load configs
	cfgs, err := LoadConfigsFromFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	// THEN

	if _, ok := cfgs["DEFAULT"]; !ok {
		t.Fatal("Expected DEFAULT config to exist in map")
	}
}

func TestNewConfigDefaultsPopulated(t *testing.T) {
	// GIVEN

	// Get DEFAULT config
	cfg, keyFile, err := BaseTestConfig()
	defer os.Remove(keyFile.Name())

	admin := cfg.Section("ADMIN")
	admin.NewKey("user", "ocid1.user.oc1..bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	admin.NewKey("fingerprint", "11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11")

	// Write test config to file
	f, err := WriteTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	// WHEN

	cfgs, err := LoadConfigsFromFile(f.Name())
	adminConfig, ok := cfgs["ADMIN"]

	// THEN

	if !ok {
		t.Fatal("Expected ADMIN config to exist in map")
	}

	if adminConfig.Region != "us-phoenix-1" {
		t.Errorf("Expected 'us-phoenix-1', got '%s'", adminConfig.Region)
	}
}

func TestNewConfigDefaultsOverridden(t *testing.T) {
	// GIVEN

	// Get DEFAULT config
	cfg, keyFile, err := BaseTestConfig()
	defer os.Remove(keyFile.Name())

	admin := cfg.Section("ADMIN")
	admin.NewKey("user", "ocid1.user.oc1..bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	admin.NewKey("fingerprint", "11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11")

	// Write test config to file
	f, err := WriteTestConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	// WHEN

	cfgs, err := LoadConfigsFromFile(f.Name())
	adminConfig, ok := cfgs["ADMIN"]

	// THEN

	if !ok {
		t.Fatal("Expected ADMIN config to exist in map")
	}

	if adminConfig.Fingerprint != "11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11" {
		t.Errorf("Expected fingerprint '11:11:11:11:11:11:11:11:11:11:11:11:11:11:11:11', got '%s'",
			adminConfig.Fingerprint)
	}
}
