package check

import (
	"fmt"
	"os"
	"path/filepath"
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

type fileGlob struct {
	filepath string
}

func (fe fileGlob) Check(_, _ string, _ error) error {
	matches, err := filepath.Glob(fe.filepath)
	if err != nil {
		return fmt.Errorf("error evaluating file glob pattern %q: %v", fe.filepath, err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no matches found for file glob pattern %q", fe.filepath)
	}

	return nil
}

func FileGlob(filename string) Checker {
	return fileGlob{
		filepath: filename,
	}
}
