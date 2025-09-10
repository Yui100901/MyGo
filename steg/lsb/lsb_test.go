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

var format = FormatPNG
var bitCount = 2

func TestLSB_EmbedAndExtract(t *testing.T) {

	// 打开载体图像
	carrierFile, err := os.Open("../testdata/hello.png")
	if err != nil {
		panic(fmt.Errorf("载体图像打开失败: %v", err))
	}
	defer carrierFile.Close()

	// 构造 payload 数据
	payload := []byte(`
隐写术是一门关于信息隐藏的技巧与科学，所谓信息隐藏指的是不让除预期的接收者之外的任何人知晓信息的传递事件或者信息的内容。隐写术的英文叫做Steganography，来源于特里特米乌斯的一本讲述密码学与隐写术的著作Steganographia，该书书名源于希腊语，意为“隐秘书写”。
一般来说，隐写的信息看起来像一些其他的东西，例如一张购物清单，一篇文章，一篇图画或者其他“伪装”（cover）的消息。
隐写的信息通常用一些传统的方法进行加密，然后用某种方法修改一个“伪装文本”（covertext），使其包含被加密过的消息，形成所谓的“隐秘文本”（stegotext）。例如，文字的大小、间距、字体，或者掩饰文本的其他特性可以被修改来包含隐藏的信息。只有接收者知道所使用的隐藏技术，才能够恢复信息，然后对其进行解密。
首先在概述隐写术时必须提到它的近亲兄弟电子水印（Watermarking），水印用于识别物品的真伪（比如：新台币上面翻转隐约可见到梅花、人民币上面的隐约可见的毛泽东头像），或者作为著作权声明的标志，或者加入作品属性信息。电子水印与隐写术的相同点是，二者都是将一个文件隐写至另一个文件当中，而两者的区别在于使用目的与处理算法的不同。电子隐写侧重将秘密文件隐藏，而电子水印则较重视著作权的声明与维护，防止多媒体作品被非法复制等等。电子隐写术一旦被识破，则秘密文件十分容易被读取，相反，电子水印并不隐藏及隐写文件的隐蔽性，而在乎加强（Robustness）除去算法的攻击。
载体文件（cover file）相对隐秘文件的大小（指数据含量，以比特计）越大，隐藏后者就越加容易。
因为这个原因，数字图像（包含有大量的数据）在因特网和其他传媒上被广泛用于隐藏消息。这种方法使用的广泛程度无从查考。例如：一个24位的位图中的每个像素的三个颜色分量（红，绿和蓝）各使用8个比特来表示。如果我们只考虑蓝色的话，就是说有2种不同的数值来表示深浅不同的蓝色。而像11111111和11111110这两个值所表示的蓝色，人眼几乎无法区分。因此，这个最低有效位就可以用来存储颜色之外的信息，而且在某种程度上几乎是检测不到的。如果对红色和绿色进行同样的操作，就可以在差不多三个像素中存储一个字节的信息。
更正式一点地说，使隐写的信息难以探测的，也就是保证“有效载荷”（需要被隐蔽的信号）对“载体”（即原始的信号）的调制对载体的影响看起来（理想状况下甚至在统计上）可以忽略。这就是说，这种改变应该无法与载体中的噪声加以区别。
（从信息论的观点来看，这就是说信道的容量必须大于传输“表面上”的信号的需求。这就叫做信道的冗余。对于一幅数字图像，这种冗余可能是成像单元的噪声；对于数字音频，可能是录音或者放大设备所产生的噪声。任何有着模拟放大级的系统都会有所谓的热噪声（或称“1/f”噪声)，这可以用作掩饰。另外，有损压缩技术（如JPEG）会在解压后的数据中引入一些误差，利用这些误差作隐写术用途也是可能的。）
隐写术也可以用作数字水印，这里一条消息（往往只是一个标识符）被隐藏到一幅图像中，使得其来源能够被跟踪或校验。应用电脑的字体设计的隐写术。
`)
	payloadReader := bytes.NewReader(payload)
	// 创建 LSB 实例，设置输出格式为 PNG
	encoder := NewLSB(format, bitCount)
	t.Log(encoder)
	// 执行嵌入
	embeddedReader, err := encoder.Embed(carrierFile, payloadReader)
	if err != nil {
		panic(fmt.Errorf("嵌入失败: %v", err))
	}

	// 保存嵌入后的图像为 embedded.png
	outFile, err := os.Create(string("embedded." + format))
	if err != nil {
		panic(fmt.Errorf("输出文件创建失败: %v", err))
	}
	defer outFile.Close()

	// 将嵌入后的图像写入文件
	_, err = outFile.ReadFrom(embeddedReader)
	if err != nil {
		panic(fmt.Errorf("写入失败: %v", err))
	}

	fmt.Println("嵌入完成，图像已保存为 embedded." + format)
}

func TestLSB_Extract(t *testing.T) {
	// 打开嵌入后的图像
	embeddedFile, err := os.Open(string("embedded." + format))
	if err != nil {
		panic(fmt.Errorf("嵌入图像打开失败: %v", err))
	}
	defer embeddedFile.Close()

	// 创建 LSB 实例
	decoder := NewLSB(format, bitCount)

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
