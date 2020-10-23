// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package iamkey

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestKey_JSONEncoding(t *testing.T) {
	data, err := json.Marshal(testKey(t))
	require.NoError(t, err)

	key := &Key{}
	err = json.Unmarshal(data, key)
	require.NoError(t, err)
	assert.Equal(t, testKey(t), key)
}

func TestKey_YAMLEncoding(t *testing.T) {
	data, err := yaml.Marshal(testKey(t))
	require.NoError(t, err)

	key := &Key{}
	err = yaml.Unmarshal(data, key)
	require.NoError(t, err)
	assert.Equal(t, testKey(t), key)
}

func TestKey_WriteFileReadFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "yc-sdk")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "key.json")
	err = WriteToJSONFile(file, testKey(t))
	require.NoError(t, err)

	keyClone, err := ReadFromJSONFile(file)
	require.NoError(t, err)
	assert.Equal(t, testKey(t), keyClone)
}

func testKey(t *testing.T) *Key {
	data, err := ioutil.ReadFile("../test_data/service_account_key.pb")
	require.NoError(t, err)
	key := &Key{}
	err = proto.UnmarshalText(string(data), key)
	require.NoError(t, err)
	return key
}
