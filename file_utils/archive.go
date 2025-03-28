package file_utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//
// @Author yfy2001
// @Date 2025/3/26 22 10
//

// CreateTarArchive 创建tar文件
// 源目标可为文件或目录
func CreateTarArchive(src, dest string) error {
	// 获取源路径的元信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法获取源路径信息: %v", err)
	}

	// 根据类型确定基准路径
	var basePath string
	if srcInfo.IsDir() {
		basePath = src
	} else {
		basePath = filepath.Dir(src)
	}

	tarFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	// 遍历文件树
	return filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("遍历文件失败: %v", err)
		}

		// 跳过源目录自身（如果是目录）
		if filePath == src && info.IsDir() {
			return nil
		}

		// 计算相对于基准路径的相对路径
		relPath, err := filepath.Rel(basePath, filePath)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %v", err)
		}
		relPath = filepath.ToSlash(relPath) // 统一为斜杠路径

		// 创建 tar 头部信息
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return fmt.Errorf("创建头部失败: %v", err)
		}
		header.Name = relPath // 关键：覆盖自动生成的路径

		// 写入头部
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("写入头部失败: %v", err)
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("打开文件失败: %v", err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("写入内容失败: %v", err)
			}
		}

		return nil
	})
}

func DecompressTar(src, dest string) error {
	// 打开tar文件
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v\n", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // 结束循环
		}
		if err != nil {
			return fmt.Errorf("tar.Next() failed: %v", err)
		}

		// 防止路径遍历漏洞
		targetPath := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(dest)) {
			return fmt.Errorf("路径不安全: %s", hdr.Name)
		}

		// 根据文件类型处理
		switch hdr.Typeflag {
		case tar.TypeDir:
			// 创建目录并设置权限
			if err := os.MkdirAll(targetPath, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("创建父目录失败: %v", err)
			}

			// 使用匿名函数包裹文件操作，确保defer及时执行
			if err := func() error {
				// 创建文件并设置权限
				f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
				if err != nil {
					return fmt.Errorf("创建文件失败: %v", err)
				}
				defer f.Close() // 此defer在匿名函数结束时触发

				// 写入文件内容
				if _, err := io.Copy(f, tr); err != nil {
					return fmt.Errorf("写入文件内容失败: %v", err)
				}

				// 显式设置权限（解决umask影响）
				if err := os.Chmod(targetPath, os.FileMode(hdr.Mode)); err != nil {
					return fmt.Errorf("设置文件权限失败: %v", err)
				}
				return nil
			}(); err != nil {
				return err // 传递内部错误
			}
		case tar.TypeSymlink:
			if err := os.Symlink(hdr.Linkname, targetPath); err != nil {
				return fmt.Errorf("创建符号链接失败: %v", err)
			}
		default:
			return fmt.Errorf("不支持的文件类型: %v in %s", hdr.Typeflag, hdr.Name)
		}
	}

	return nil
}

// CreateZipArchive 创建zip文件
// 源目标可为文件或目录
func CreateZipArchive(src, dest string) error {
	// 创建zip文件
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建zip写入器
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 遍历目录并添加到zip中
	err = filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(src, filePath)
		if err != nil {
			return err
		}

		// 如果是目录，跳过（zip文件会自动处理目录结构）
		if info.IsDir() {
			return nil
		}

		// 创建zip文件的条目
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// 打开文件并写入zip条目
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFile, file)
		return err
	})

	return err
}

// DecompressZip 解压 ZIP 文件到指定目录
func DecompressZip(src, dest string) error {
	// 打开 ZIP 文件
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close() // 在函数结束时释放资源

	// 遍历 ZIP 文件中的每个文件
	for _, file := range r.File {
		if err := func() error {
			fpath := filepath.Join(dest, filepath.FromSlash(file.Name))
			if file.FileInfo().IsDir() {
				// 如果是目录，则创建目录
				if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
					return err
				}
				return nil
			}

			// 确保目标文件的目录存在
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			// 打开压缩文件内容
			inFile, err := file.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			// 创建目标文件
			outFile, err := os.Create(fpath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// 复制内容
			_, err = io.Copy(outFile, inFile)
			return err
		}(); err != nil {
			return err
		}
	}
	return nil
}

// CreateGzipArchive 创建gzip文件
// 源目标只能为文件，一般与tar一起使用，归档为tar.gz
func CreateGzipArchive(src, dest string) error {
	// 打开源文件
	inFile, err := os.Open(src)
	if err != nil {
		inFile.Close()
		return err
	}

	// 创建 Gzip 文件
	outFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 创建 Gzip Writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// 将文件内容写入 Gzip 文件
	_, err = io.Copy(gzipWriter, inFile)
	if err != nil {
		return err
	}

	return nil
}

// DecompressGzip 解压 gzip 文件
func DecompressGzip(src, dest string) error {
	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()

	gzipReader, err := gzip.NewReader(inFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	outFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gzipReader)
	return err
}
