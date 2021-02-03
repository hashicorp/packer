package command

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func mustString(s string, e error) string {
	if e != nil {
		panic(e)
	}
	return s
}

func createFiles(dir string, content map[string]string) {
	for relPath, content := range content {
		contentPath := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(contentPath), 0777); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(contentPath, []byte(content), 0666); err != nil {
			panic(err)
		}
		log.Printf("created tmp file: %s", contentPath)
	}
}

type configDirSingleton struct {
	dirs map[string]string
}

// when you call dir twice with the same key, the result should be the same
func (c *configDirSingleton) dir(key string) string {
	if v, exists := c.dirs[key]; exists {
		return v
	}
	c.dirs[key] = mustString(ioutil.TempDir("", "pkr-test-cfg-dir-"+key))
	return c.dirs[key]
}
