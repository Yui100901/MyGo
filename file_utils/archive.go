package file_utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

//
// @Author yfy2001
// @Date 2025/3/26 22 10
//

// CreateTarArchive 创建tar文件
// 源目标可为文件或目录
func CreateTarArchive(src, dest string) error {
	tarFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	// 遍历目录并将内容写入 tar 文件
	err = filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过源目录本身
		if filePath == src {
			return nil
		}

		// 创建 tar 头信息
		header, err := tar.FileInfoHeader(info, filePath)
		if err != nil {
			return err
		}

		// 设置归档文件的相对路径
		relativePath, err := filepath.Rel(src, filePath)
		header.Name = filepath.ToSlash(relativePath)

		// 写入头信息
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件，写入文件数据
		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
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
		fpath := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			// 如果是目录，则创建目录
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
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

		// 创建目标文件
		outFile, err := os.Create(fpath)
		if err != nil {
			inFile.Close() // 及时关闭 inFile
			return err
		}

		// 将内容复制到目标文件
		_, err = io.Copy(outFile, inFile)
		// 显式关闭文件
		inFile.Close()
		outFile.Close()

		if err != nil {
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
		return err
	}
	defer inFile.Close()

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
