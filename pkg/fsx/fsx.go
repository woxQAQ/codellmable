package fsx

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	constIndent        = "│   "
	constConnector     = "├── "
	constLastConnector = "└── "
)

func Tree(dirPath string, exclude ...string) ([]byte, error) {
	if !filepath.IsAbs(dirPath) {
		var err error
		dirPath, err = filepath.Abs(dirPath)
		if err != nil {
			return nil, err
		}
	}
	buf := bytes.Buffer{}
	rootDepth := strings.Count(dirPath, string(os.PathSeparator))
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relRoot, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		if info.IsDir() && slices.Contains(exclude, relRoot) {
			return filepath.SkipDir
		}
		depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
		indent := strings.Repeat(constIndent, depth)

		connector := constConnector

		if info.IsDir() && path != dirPath {
			connector = constLastConnector
		}

		_, err = buf.WriteString(fmt.Sprintf("%s%s%s\n", indent, connector, info.Name()))
		return err
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
