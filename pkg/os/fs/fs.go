package fs

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	DirectoryPerm      os.FileMode = 0755
	RegularFilePerm    os.FileMode = 0644
	ExecutableFilePerm os.FileMode = 0755
)

// NormalizePath returns the string path replaced from \(windows) to /(unix)
func NormalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// Join joins any number of path elements into a single path
func Join(paths ...string) string {
	return NormalizePath(filepath.Join(paths...))
}

// Children returns children of path
func Children(path string) string {
	return path[strings.Index(path, "/")+1:]
}

// IsFile returns true when path is file, false otherwise
func IsFile(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !stat.IsDir(), nil
}

// IsDir returns true when path is directory, false otherwise
func IsDir(path string) (bool, error) {
	isFile, err := IsFile(path)
	if err != nil {
		return false, err
	}
	return !isFile, nil
}

// WriteFile create and truncate a file after write data from
// dataSource, the not exists directories will be auto created
func WriteFile(filename string, dataSource io.Reader, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), DirectoryPerm); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if dataSource != nil {
		_, err = io.Copy(file, dataSource)
		return err
	}

	return nil
}
