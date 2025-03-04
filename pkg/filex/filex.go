package filex

import (
	"log"
	"os"
	"path/filepath"
)

const dirPermission = 0755

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return os.IsExist(err)
}

func MustGetAbsPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal("Path conversion failed:", err)
	}
	return absPath
}

func PrepareOutputDir(path string) error {
	return os.MkdirAll(path, dirPermission)
}
