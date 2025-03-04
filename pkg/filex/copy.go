package filex

import (
	"io"
	"os"
	"path/filepath"
)

func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == src {
			return nil
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return createDir(dstPath, info)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, dstPath)
		}
		return copyFile(path, dstPath, info)
	})
}

func createDir(dstPath string, info os.FileInfo) error {
	if err := os.MkdirAll(dstPath, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Chtimes(dstPath, info.ModTime(), info.ModTime())
}

func copyFile(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(target, dst)
}
