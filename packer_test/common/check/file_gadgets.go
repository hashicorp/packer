package check

import (
	"fmt"
	"os"
	"regexp"
)

type fileExists struct {
	filepath string
	isDir    bool
}

func (fe fileExists) Check(_, _ string, _ error) error {
	st, err := os.Stat(fe.filepath)
	if err != nil {
		return fmt.Errorf("failed to stat %q: %s", fe.filepath, err)
	}

	if st.IsDir() && !fe.isDir {
		return fmt.Errorf("file %q is a directory, wasn't supposed to be", fe.filepath)
	}

	if !st.IsDir() && fe.isDir {
		return fmt.Errorf("file %q is not a directory, was supposed to be", fe.filepath)
	}

	return nil
}

func FileExists(filePath string, isDir bool) Checker {
	return fileExists{
		filepath: filePath,
		isDir:    isDir,
	}
}

type fileInDir struct {
	filename string
	dirPath  string
}

func (fe fileInDir) Check(_, _ string, _ error) error {
	files, err := os.ReadDir(fe.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read dir %q: %s", fe.dirPath, err)
	}

	pattern := regexp.MustCompile(fe.filename)

	for _, file := range files {
		if !file.IsDir() && pattern.MatchString(file.Name()) {
			return nil
		}
	}

	return fmt.Errorf("file %q not found in dir %q", fe.filename, fe.dirPath)
}

func FileInDir(dirPath string, filename string) Checker {
	return fileInDir{
		filename: filename,
		dirPath:  dirPath,
	}
}
