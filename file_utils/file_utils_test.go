package file_utils

import (
	"os"
	"testing"
)

//
// @Author yfy2001
// @Date 2025/3/24 12 44
//

// 测试 CreateTarArchive 函数
func TestCreateTarArchive(t *testing.T) {
	src := "./testdata"      // 源文件夹
	dest := "./testdata.tar" // 目标 TAR 文件

	// 创建 TAR 文件
	err := CreateTarArchive(src, dest)
	if err != nil {
		t.Errorf("CreateTarArchive failed: %v", err)
	}

	// 验证目标文件是否存在
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("TAR file not created: %v", err)
	}

	// 清理测试数据
	//os.Remove(dest)
}

// 测试 CreateZipArchive 函数
func TestCreateZipArchive(t *testing.T) {
	src := "./testdata"      // 源文件夹
	dest := "./testdata.zip" // 目标 ZIP 文件

	// 创建 ZIP 文件
	err := CreateZipArchive(src, dest)
	if err != nil {
		t.Errorf("CreateZipArchive failed: %v", err)
	}

	// 验证目标文件是否存在
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("ZIP file not created: %v", err)
	}

	// 清理测试数据
	//os.Remove(dest)
}

// 测试 DecompressGzip 函数
func TestDecompressGzip(t *testing.T) {
	src := "./testdata.gz" // 示例 GZIP 文件（需要事先准备）
	dest := "./output.txt" // 解压后的文件

	// 解压 GZIP 文件
	err := DecompressGzip(src, dest)
	if err != nil {
		t.Errorf("DecompressGzip failed: %v", err)
	}

	// 验证解压后的文件是否存在
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("Decompressed file not created: %v", err)
	}

	// 清理测试数据
	os.Remove(dest)
}
