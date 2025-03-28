package file_utils

import (
	"fmt"
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

func TestDecompressTarArchive(t *testing.T) {
	srcFile := "testdata.tar"
	dstDir := "./testdata-tar-extracted"

	// 打开tar文件
	file, err := os.Open(srcFile)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer file.Close()

	// 解压tar文件
	if err := DecompressTar(dstDir, file); err != nil {
		fmt.Printf("解压失败: %v\n", err)
	} else {
		fmt.Println("解压成功！")
	}
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

func TestDecompressZip(t *testing.T) {
	src := "./testdata.zip" // 示例 ZIP 文件（需要事先准备）
	dest := "./testdatazip" // 解压后的文件
	// 解压 GZIP 文件
	err := DecompressZip(src, dest)
	if err != nil {
		t.Errorf("DecompressZip failed: %v", err)
	}

	// 验证解压后的文件是否存在
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("DecompressZip file not created: %v", err)
	}

	// 清理测试数据
	//os.Remove(dest)
}

// 测试 DecompressGzip 函数
func TestDecompressGzip(t *testing.T) {
	src := "./testdata.tar.gz" // 示例 GZIP 文件（需要事先准备）
	dest := "./testdatatargz"  // 解压后的文件

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
	//os.Remove(dest)
}

func TestCreateGzipArchive(t *testing.T) {
	src := "./testdata.tar"     // 源文件夹
	dest := "./testdata.tar.gz" // 目标 GZIP 文件

	// 创建 gZIP 文件
	err := CreateGzipArchive(src, dest)
	if err != nil {
		t.Errorf("CreateGzipArchive failed: %v", err)
	}

	// 验证目标文件是否存在
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Errorf("GZIP file not created: %v", err)
	}

	// 清理测试数据
	//os.Remove(dest)
}
