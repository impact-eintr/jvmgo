package classpath

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type ZipEntry struct {
	absPath string
	zipRC *zip.ReadCloser
}

func newZipEntry(path string) *ZipEntry {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &ZipEntry{absPath: absPath, zipRC: nil}
}

func (self *ZipEntry) readClass(className string) ([]byte, Entry, error) {
	if self.zipRC == nil {
		err := self.openJar()
		if err != nil {
			return nil, nil, errors.New("class not found: " + className)
		}
	}

	classFile := self.findClass(className)
	if classFile == nil {
		return nil, nil, errors.New("class not found: " + className)
	}

	data, err := readClass(classFile)
	return data, self, err
}

func (self *ZipEntry) openJar() error {
	r, err := zip.OpenReader(self.absPath)
	if err == nil {
		self.zipRC = r
	}
	return err
}

func (self *ZipEntry) findClass(className string) *zip.File {
	for _, f := range self.zipRC.File {
		if f.Name == className {
			return f
		}
	}
	return nil
}

func readClass(classFile *zip.File) ([]byte, error) {
	rc, err := classFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// read class data
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (self *ZipEntry) String() string {
	return self.absPath
}
