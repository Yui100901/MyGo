package main

import (
	"fmt"
	"os"
	"testing"
)

//
// @Author yfy2001
// @Date 2025/9/10 11 30
//

func TestBind(t *testing.T) {
	carrier, _ := os.Open("hello.png")
	defer carrier.Close()
	payload, _ := os.Open("hello.zip")
	defer payload.Close()

	binder := NewBind("png")
	embed, err := binder.Embed(carrier, payload)
	if err != nil {
		return
	}
	// 保存嵌入后的图像为 embedded.png
	outFile, err := os.Create("embedded." + binder.carrierFormat)
	if err != nil {
		panic(fmt.Errorf("输出文件创建失败: %v", err))
	}
	defer outFile.Close()

	// 将嵌入后的图像写入文件
	_, err = outFile.ReadFrom(embed)
	if err != nil {
		panic(fmt.Errorf("写入失败: %v", err))
	}
}
