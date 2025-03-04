package filex

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCopyDir(t *testing.T) {
	// 创建临时测试目录
	testCases := []struct {
		name        string
		setup       func(string) // 测试目录初始化函数
		wantErr     bool         // 是否期望错误
		skipWindows bool         // 是否跳过Windows测试
	}{
		{
			name: "empty directory",
			setup: func(src string) {
				// 空目录不需要额外操作
			},
		},
		{
			name: "nested directories",
			setup: func(src string) {
				createTestFile(t, src, "file1.txt", "content1", 0644)
				createTestDir(t, src, "subdir", 0755)
				createTestFile(t, src, "subdir/file2.txt", "content2", 0600)
				createTestFile(t, src, "subdir/file3.txt", "content3", 0666)
			},
		},
		{
			name: "special permissions",
			setup: func(src string) {
				createTestFile(t, src, "readonly.txt", "ro-content", 0400)
				createTestDir(t, src, "writabledir", 0777)
			},
		},
		{
			name:    "nonexistent source",
			setup:   func(src string) {},
			wantErr: true,
		},
		{
			name: "permission denied",
			setup: func(src string) {
				if runtime.GOOS != "windows" {
					createTestFile(t, src, "unreadable.txt", "secret", 0000)
				}
			},
			wantErr:     true,
			skipWindows: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipWindows && runtime.GOOS == "windows" {
				t.Skip("Skipping test on Windows")
			}

			// 准备测试环境
			src := t.TempDir()
			dst := filepath.Join(t.TempDir(), "copy-dest")

			// 特殊处理不存在的源目录测试
			if tc.name == "nonexistent source" {
				src = filepath.Join(t.TempDir(), "nonexistent")
			} else {
				tc.setup(src)
			}

			// 执行拷贝操作
			err := CopyDir(src, dst)

			// 错误验证
			if (err != nil) != tc.wantErr {
				t.Fatalf("CopyDir() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return // 不需要后续验证
			}

			// 目录结构验证
			verifyDirectoryTree(t, src, dst)

			// 元数据验证
			verifyMetadata(t, src, dst)
		})
	}
}

// 创建测试文件辅助函数
func createTestFile(t *testing.T, base string, path string, content string, perm fs.FileMode) {
	fullPath := filepath.Join(base, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
	// 设置修改时间为固定值便于验证
	if err := os.Chtimes(fullPath, time.Now(), time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

// 创建测试目录辅助函数
func createTestDir(t *testing.T, base string, path string, perm fs.FileMode) {
	fullPath := filepath.Join(base, path)
	if err := os.MkdirAll(fullPath, perm); err != nil {
		t.Fatal(err)
	}
}

// 目录结构验证函数
func verifyDirectoryTree(t *testing.T, src string, dst string) {
	srcMap := make(map[string]fs.FileInfo)
	dstMap := make(map[string]fs.FileInfo)

	// 遍历源目录
	filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		srcMap[relPath] = info
		return nil
	})

	// 遍历目标目录
	filepath.Walk(dst, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(dst, path)
		dstMap[relPath] = info
		return nil
	})

	// 比较目录结构
	for path, srcInfo := range srcMap {
		dstInfo, exists := dstMap[path]
		if !exists {
			t.Errorf("Missing in destination: %s", path)
			continue
		}

		// 类型检查
		if srcInfo.IsDir() != dstInfo.IsDir() {
			t.Errorf("Type mismatch for %s: src is dir %v, dst is dir %v",
				path, srcInfo.IsDir(), dstInfo.IsDir())
		}
	}

	// 检查多余文件
	for path := range dstMap {
		if _, exists := srcMap[path]; !exists {
			t.Errorf("Extra file in destination: %s", path)
		}
	}
}

// 元数据验证函数
func verifyMetadata(t *testing.T, src string, dst string) {
	filepath.Walk(src, func(srcPath string, srcInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, srcPath)
		dstPath := filepath.Join(dst, relPath)

		// 跳过根目录
		if relPath == "." {
			return nil
		}

		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Errorf("Failed to stat destination file: %s", dstPath)
			return nil
		}

		// 权限验证
		if srcInfo.Mode().Perm() != dstInfo.Mode().Perm() {
			t.Errorf("Permission mismatch for %s:\nsrc: %04o\ndst: %04o",
				relPath, srcInfo.Mode().Perm(), dstInfo.Mode().Perm())
		}

		// 修改时间验证（允许1秒误差）
		if !srcInfo.ModTime().Round(time.Second).Equal(dstInfo.ModTime().Round(time.Second)) {
			t.Errorf("ModTime mismatch for %s:\nsrc: %v\ndst: %v",
				relPath, srcInfo.ModTime(), dstInfo.ModTime())
		}

		// 文件内容验证
		if !srcInfo.IsDir() {
			srcContent, _ := os.ReadFile(srcPath)
			dstContent, _ := os.ReadFile(dstPath)
			if string(srcContent) != string(dstContent) {
				t.Errorf("Content mismatch for %s", relPath)
			}
		}

		return nil
	})
}
