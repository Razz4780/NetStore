package internal

import (
	"io/ioutil"
	"os"
)

const (
	ReceivedFilesDir = "tmp"
)

func CreateReceivedFilesDir() error {
	if _, err := os.Stat(ReceivedFilesDir); os.IsNotExist(err) {
		return os.Mkdir(ReceivedFilesDir, os.ModeDir)
	} else {
		return err
	}
}

func OpenFile(filename string, offset int64, flag int) (*os.File, error) {
	file, err := os.OpenFile(filename, flag, 0666)
	if err != nil {
		return nil, err
	}
	if offset != 0 {
		if _, err := file.Seek(offset, 0); err != nil {
			return nil, err
		}
	}
	return file, nil
}

func IndexFiles(dir string) ([]os.FileInfo, error) {
	allFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	regFiles := make([]os.FileInfo, 0, len(allFiles))
	for _, file := range allFiles {
		if file.Mode().IsRegular() {
			regFiles = append(regFiles, file)
		}
	}
	return regFiles, nil
}
