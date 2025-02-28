package converter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

//
// @Author yfy2001
// @Date 2025/2/28 13 26
//

type RGB struct {
	R, G, B int
}

type HEX struct {
	Value string
}

type HSV struct {
	H, S, V float64
}

type CMYK struct {
	C, M, Y, K float64
}

// Color 是主结构体，包含嵌套的各种颜色类型
type Color struct {
	RGB  RGB
	HEX  HEX
	HSV  HSV
	CMYK CMYK
}

// NewFromRGB 使用 RGB 初始化 Color，并自动转换为其他颜色类型
func NewFromRGB(r, g, b int) *Color {
	color := &Color{}
	color.RGB = RGB{R: r, G: g, B: b}
	color.HEX = RGBToHEX(color.RGB)
	color.HSV = RGBToHSV(color.RGB)
	color.CMYK = RGBToCMYK(color.RGB)
	return color
}

// NewFromHEX 使用 HEX 初始化 Color，并自动转换为其他颜色类型
func NewFromHEX(hex string) (*Color, error) {
	color := &Color{}
	rgb, err := HEXToRGB(HEX{Value: hex})
	if err != nil {
		return nil, err
	}
	color.RGB = rgb
	color.HEX = HEX{Value: hex}
	color.HSV = RGBToHSV(color.RGB)
	color.CMYK = RGBToCMYK(color.RGB)
	return color, nil
}

// NewFromHSV 使用 HSV 初始化 Color，并自动转换为其他颜色类型
func NewFromHSV(h, s, v float64) *Color {
	color := &Color{}
	color.HSV = HSV{H: h, S: s, V: v}
	color.RGB = HSVToRGB(color.HSV)
	color.HEX = RGBToHEX(color.RGB)
	color.CMYK = RGBToCMYK(color.RGB)
	return color
}

// NewFromCMYK 使用 CMYK 初始化 Color，并自动转换为其他颜色类型
func NewFromCMYK(c, m, y, k float64) *Color {
	color := &Color{}
	color.CMYK = CMYK{C: c, M: m, Y: y, K: k}
	color.RGB = CMYKToRGB(color.CMYK)
	color.HEX = RGBToHEX(color.RGB)
	color.HSV = RGBToHSV(color.RGB)
	return color
}

// RGBToHEX Convert RGB to HEX
func RGBToHEX(rgb RGB) HEX {
	return HEX{
		Value: fmt.Sprintf("#%02X%02X%02X", rgb.R, rgb.G, rgb.B),
	}
}

// HEXToRGB Convert HEX to RGB
func HEXToRGB(hex HEX) (RGB, error) {
	value := hex.Value
	if strings.HasPrefix(value, "#") {
		value = value[1:]
	}
	if len(value) != 6 {
		return RGB{}, fmt.Errorf("invalid hex color format")
	}

	r, err := strconv.ParseInt(value[0:2], 16, 32)
	if err != nil {
		return RGB{}, err
	}
	g, err := strconv.ParseInt(value[2:4], 16, 32)
	if err != nil {
		return RGB{}, err
	}
	b, err := strconv.ParseInt(value[4:6], 16, 32)
	if err != nil {
		return RGB{}, err
	}

	return RGB{R: int(r), G: int(g), B: int(b)}, nil
}

// RGBToHSV Convert RGB to HSV
func RGBToHSV(rgb RGB) HSV {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	maxValue := math.Max(r, math.Max(g, b))
	minValue := math.Min(r, math.Min(g, b))
	delta := maxValue - minValue

	var h float64
	if delta == 0 {
		h = 0
	} else if maxValue == r {
		h = math.Mod((g-b)/delta, 6)
	} else if maxValue == g {
		h = (b-r)/delta + 2
	} else {
		h = (r-g)/delta + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}

	s := 0.0
	if maxValue != 0 {
		s = delta / maxValue
	}

	v := maxValue

	return HSV{H: h, S: s, V: v}
}

// HSVToRGB Convert HSV to RGB
func HSVToRGB(hsv HSV) RGB {
	h := hsv.H
	s := hsv.S
	v := hsv.V

	chroma := v * s
	x := chroma * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - chroma

	var r, g, b float64
	switch {
	case h >= 0 && h < 60:
		r, g, b = chroma, x, 0
	case h >= 60 && h < 120:
		r, g, b = x, chroma, 0
	case h >= 120 && h < 180:
		r, g, b = 0, chroma, x
	case h >= 180 && h < 240:
		r, g, b = 0, x, chroma
	case h >= 240 && h < 300:
		r, g, b = x, 0, chroma
	case h >= 300 && h < 360:
		r, g, b = chroma, 0, x
	}

	return RGB{
		R: int((r + m) * 255),
		G: int((g + m) * 255),
		B: int((b + m) * 255),
	}
}

// RGBToCMYK Convert RGB to CMYK
func RGBToCMYK(rgb RGB) CMYK {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	k := 1 - math.Max(r, math.Max(g, b))
	var c, m, y float64
	if k < 1 {
		c = (1 - r - k) / (1 - k)
		m = (1 - g - k) / (1 - k)
		y = (1 - b - k) / (1 - k)
	}

	return CMYK{C: c, M: m, Y: y, K: k}
}

// CMYKToRGB Convert CMYK to RGB
func CMYKToRGB(cmyk CMYK) RGB {
	r := (1 - cmyk.C) * (1 - cmyk.K)
	g := (1 - cmyk.M) * (1 - cmyk.K)
	b := (1 - cmyk.Y) * (1 - cmyk.K)

	return RGB{
		R: int(math.Round(r * 255)), // Use math.Round to avoid precision issues
		G: int(math.Round(g * 255)), // Round to the nearest integer
		B: int(math.Round(b * 255)), // Round to the nearest integer
	}
}
