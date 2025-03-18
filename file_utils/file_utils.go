package file_utils

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

// FileData 结构体定义
type FileData struct {
	Path     string
	PathBuf  string
	AbsPath  string
	Filename string
	Metadata os.FileInfo
}

// NewFileData 函数，用于创建 FileData 实例
func NewFileData(path string) (*FileData, error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("can't canonicalize path %s: %v", path, err)
	}

	// 获取文件名
	filename := filepath.Base(absPath)

	// 获取文件元数据
	metadata, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	return &FileData{
		Path:     path,
		PathBuf:  absPath,
		AbsPath:  absPath,
		Filename: filename,
		Metadata: metadata,
	}, nil
}

// TraverseDirFiles 函数，遍历给定目录并返回文件路径列表
// `recursive` 参数表明是否递归遍历子目录
func TraverseDirFiles(dir string, recursive bool) ([]*FileData, []*FileData, error) {
	var files []*FileData
	var dirs []*FileData

	dirFileData, err := NewFileData(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open directory: %v", err)
	}

	entries, err := os.ReadDir(dirFileData.PathBuf)
	if err != nil {
		return nil, nil, err
	}

	for _, entry := range entries {
		path, err := filepath.EvalSymlinks(filepath.Join(dirFileData.PathBuf, entry.Name()))
		if err != nil {
			continue
		}
		data, err := NewFileData(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get file data: %v", err)
		}

		if data.Metadata.IsDir() {
			dirs = append(dirs, data)
			if recursive {
				subDirs, subFiles, err := TraverseDirFiles(path, true)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to traverse subdirectory: %v", err)
				}
				dirs = append(dirs, subDirs...)
				files = append(files, subFiles...)
			}
		} else {
			files = append(files, data)
		}
	}

	return files, dirs, nil
}

// Replace 函数，用于替换源文件到目标文件
func Replace(source, target string) (string, error) {
	_, err := os.Stat(source)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("source file %s does not exist", source)
	}

	_, err = os.Stat(target)
	if err == nil {
		os.Remove(target)
	}

	input, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(target, input, 0644)
	if err != nil {
		return "", err
	}

	return "文件替换成功！", nil
}

// CreateDirectory 函数，用于创建文件夹
func CreateDirectory(path string) (*FileData, error) {
	dir := filepath.Clean(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	}
	return NewFileData(dir)
}

func AddFileToZip(zipWriter *zip.Writer, filename, content string) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建 zip 归档: %v", err)
	}
	_, err = writer.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("无法写入 zip 归档: %v", err)
	}
	return nil
}
