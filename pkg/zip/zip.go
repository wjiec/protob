package zip

import (
	"archive/zip"
	"bytes"
	"io"
)

type File = zip.File

// VisitFiles walks the file tree on zipped roots, calling walker for each file
func VisitFiles(content []byte, walker func(*File) error) error {
	rd, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return err
	}

	for _, f := range rd.File {
		if !f.FileInfo().IsDir() {
			if err := walker(f); err != nil {
				return err
			}
		}
	}
	return nil
}

// AsReader opens the zipped file as io.Reader and calling fn to do
func AsReader(file *File, fn func(io.Reader) error) error {
	if rd, err := file.Open(); err != nil {
		return err
	} else {
		defer func() { _ = rd.Close() }()
		return fn(rd)
	}
}
