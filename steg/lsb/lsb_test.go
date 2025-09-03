package lsb

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

//
// @Author yfy2001
// @Date 2025/9/3 22 16
//

func TestLSB_EmbedAndExtract(t *testing.T) {

	// 打开载体图像
	carrierFile, err := os.Open("../testdata/hello.png")
	if err != nil {
		panic(fmt.Errorf("载体图像打开失败: %v", err))
	}
	defer carrierFile.Close()

	// 构造 payload 数据
	payload := []byte("Go语言（又称 Golang）由 Google 于 2009 年正式发布，其设计者包括三位计算机领域的传奇人物：Robert Griesemer、Rob Pike 和 Ken Thompson。他们在面对 Google 内部庞大而复杂的代码库时，意识到现有语言（如 C++、Java）在编译速度、并发处理和可维护性方面存在诸多瓶颈。因此，他们决定打造一门既高效又易用的现代语言——Go。\n\nGo 的目标很明确：为构建大型系统而生，同时保持简洁和高性能。")
	payloadReader := bytes.NewReader(payload)

	// 创建 LSB 实例，设置输出格式为 PNG
	encoder := &LSB{exportFormat: FormatPNG}

	// 执行嵌入
	embeddedReader, err := encoder.Embed(carrierFile, payloadReader)
	if err != nil {
		panic(fmt.Errorf("嵌入失败: %v", err))
	}

	// 保存嵌入后的图像为 embedded.png
	outFile, err := os.Create("embedded.png")
	if err != nil {
		panic(fmt.Errorf("输出文件创建失败: %v", err))
	}
	defer outFile.Close()

	// 将嵌入后的图像写入文件
	_, err = outFile.ReadFrom(embeddedReader)
	if err != nil {
		panic(fmt.Errorf("写入失败: %v", err))
	}

	fmt.Println("嵌入完成，图像已保存为 embedded.png")
}

func TestLSB_Extract(t *testing.T) {
	// 打开嵌入后的图像
	embeddedFile, err := os.Open("embedded.png")
	if err != nil {
		panic(fmt.Errorf("嵌入图像打开失败: %v", err))
	}
	defer embeddedFile.Close()

	// 创建 LSB 实例
	decoder := &LSB{}

	// 执行提取
	extractedReader, err := decoder.Extract(embeddedFile)
	if err != nil {
		panic(fmt.Errorf("提取失败: %v", err))
	}

	// 读取提取出的数据
	extractedData := new(bytes.Buffer)
	_, err = extractedData.ReadFrom(extractedReader)
	if err != nil {
		panic(fmt.Errorf("读取提取数据失败: %v", err))
	}

	// 打印提取结果
	fmt.Printf("提取出的数据: %s\n", extractedData.String())
}
