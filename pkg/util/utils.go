package util

import (
	"io"
	"os"
)

func ReadTxtFile(url string) ([]byte, error) {
	file, err := os.Open(url)
	if err != nil {
		return []byte(""), err
	}
	defer file.Close()
	stringData, err := io.ReadAll(file)
	if err != nil {
		return []byte(""), err
	}
	return stringData, nil
}

// write file
func WriteFileWithNosec(pathName string, data []byte) error {
	// #nosec G306, Expect WriteFile permissions to be 0600 or less
	return os.WriteFile(pathName, data, 0644)
}

func CreateDirIfNotExists(dirLocation string) error {
	if _, err := os.Stat(dirLocation); os.IsNotExist(err) {
		return os.MkdirAll(dirLocation, os.ModeDir|0755)
	}
	return nil
}

func DeleteDirs(filePath string) error {
	return os.RemoveAll(filePath)
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}
