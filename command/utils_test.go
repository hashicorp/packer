package command

import (
	"io/ioutil"
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
		if err := ioutil.WriteFile(contentPath, []byte(content), 0666); err != nil {
			panic(err)
		}
	}
}
