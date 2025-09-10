package lsb

import (
	"bytes"
	"errors"
	"github.com/Yui100901/MyGo/bit_utils"
	"golang.org/x/image/bmp"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
)

//
// @Author yfy2001
// @Date 2025/8/26 17 20
//

type ImageFormat string

const (
	FormatPNG  ImageFormat = "png"
	FormatBMP  ImageFormat = "bmp"
	FormatTIFF ImageFormat = "tiff"
)

type LSB struct {
	exportFormat ImageFormat
	writeBit     int
}

func NewLSB(format ImageFormat, bitCount int) *LSB {
	return &LSB{
		exportFormat: format,
		writeBit:     bitCount,
	}
}

// Embed 将 payload 数据嵌入 carrier 图片中，返回新的 PNG 图片数据
func (l *LSB) Embed(carrier io.Reader, payload io.Reader) (io.Reader, error) {
	// 1. 解码载体图像为 image.Image 对象
	img, _, err := image.Decode(carrier)
	if err != nil {
		return nil, err
	}

	// 2. 读取 payload 数据为字节数组
	payloadBytes, err := io.ReadAll(payload)
	if err != nil {
		return nil, err
	}

	payloadBitArr := bit_utils.NewBitArrayFromBytes(payloadBytes)

	// 3. 检查图像是否有足够容量嵌入 payload
	bounds := img.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()               // 总像素数
	if payloadBitArr.BitLen() > totalPixels*3*l.writeBit { // 每像素可嵌入 3*写入位数 个 bit（RGB）
		return nil, errors.New("载荷太大，无法嵌入图像")
	}

	// 4. 定义嵌入 bit 的函数
	embedBit := func(c byte, index *int) byte {
		embedBitArr, err := payloadBitArr.SliceBit(*index, *index+l.writeBit)
		if err != nil {
			return c // 如果所有 bit 已嵌入，返回原值
		}
		bits, _ := embedBitArr.GetByte(0)
		// 高位对齐移动到低位
		bits = bits >> (8 - l.writeBit)
		// 将该 bit 嵌入颜色分量的最低位
		mask := byte(0xFF << l.writeBit)
		c = (c & mask) | bits
		*index += l.writeBit
		return c
	}

	// 5. 创建新图像用于嵌入数据
	outImg := image.NewRGBA(bounds)
	bitIndex := 0 // 当前嵌入的 bit 索引

	// 6. 遍历每个像素，嵌入 payload 数据
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()                   // 获取原始像素的 RGBA 值（16 位）
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8) // 转换为 8 位

			// 将 payload 的 bit 嵌入 RGB 分量
			r8 = embedBit(r8, &bitIndex)
			g8 = embedBit(g8, &bitIndex)
			b8 = embedBit(b8, &bitIndex)

			// 设置新图像的像素值
			outImg.Set(x, y, color.RGBA{R: r8, G: g8, B: b8, A: uint8(a >> 8)})
		}
	}

	return EncodeImage(outImg, l.exportFormat)
}

// Extract 从 carrier 图像中提取 payload 数据
func (l *LSB) Extract(carrier io.Reader) (io.Reader, error) {
	// 1. 解码图像
	img, _, err := image.Decode(carrier)
	if err != nil {
		return nil, err
	}

	// 2. 遍历图像像素，提取 RGB 分量的最低位
	bounds := img.Bounds()
	var bitArrList []*bit_utils.BitArray

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			mask := byte((1 << l.writeBit) - 1)
			//log.Println(x, y, r, g, b, mask)
			rBit, _ := bit_utils.NewBitArray(l.writeBit)
			rBit.SetByte(0, (byte(r)&mask)<<(8-l.writeBit))
			gBit, _ := bit_utils.NewBitArray(l.writeBit)
			gBit.SetByte(0, (byte(g)&mask)<<(8-l.writeBit))
			bBit, _ := bit_utils.NewBitArray(l.writeBit)
			bBit.SetByte(0, (byte(b)&mask)<<(8-l.writeBit))
			bitArrList = append(bitArrList, rBit, gBit, bBit)
		}
	}

	// 3. 将 bit 数组转换为字节数组
	payload, _ := bit_utils.Concat(bitArrList...)
	return bytes.NewReader(payload.ToBytes()), nil // 返回提取出的 payload 数据流
}

// EncodeImage 根据指定格式将图像编码为 PNG、BMP、TIFF
func EncodeImage(img image.Image, format ImageFormat) (io.Reader, error) {
	var buf bytes.Buffer
	switch format {
	case FormatPNG:
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	case FormatBMP:
		err := bmp.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	case FormatTIFF:
		err := tiff.Encode(&buf, img, nil) // 可选参数可设置压缩方式
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(string("不支持的输出格式: " + format))
	}
	return &buf, nil
}
