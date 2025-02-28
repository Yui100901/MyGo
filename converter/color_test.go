package converter

import (
	"math"
	"testing"
)

//
// @Author yfy2001
// @Date 2025/2/28 14 02
//

// Test RGBToHEX
func TestRGBToHEX(t *testing.T) {
	rgb := RGB{R: 255, G: 99, B: 71}
	expectedHex := "#FF6347"

	hex := RGBToHEX(rgb)
	if hex.Value != expectedHex {
		t.Errorf("RGBToHEX failed: expected %s, got %s", expectedHex, hex.Value)
	}
}

// Test HEXToRGB
func TestHEXToRGB(t *testing.T) {
	hex := HEX{Value: "#FF6347"}
	expectedRGB := RGB{R: 255, G: 99, B: 71}

	rgb, err := HEXToRGB(hex)
	if err != nil {
		t.Errorf("HEXToRGB failed: unexpected error %v", err)
	}

	if rgb != expectedRGB {
		t.Errorf("HEXToRGB failed: expected %+v, got %+v", expectedRGB, rgb)
	}
}

// Test RGBToHSV
func TestRGBToHSV(t *testing.T) {
	rgb := RGB{R: 255, G: 99, B: 71}
	expectedHSV := HSV{H: 9.130434782608695, S: 0.7215686274509804, V: 1.0}

	hsv := RGBToHSV(rgb)
	if !compareFloat(hsv.H, expectedHSV.H) || !compareFloat(hsv.S, expectedHSV.S) || !compareFloat(hsv.V, expectedHSV.V) {
		t.Errorf("RGBToHSV failed: expected %+v, got %+v", expectedHSV, hsv)
	}
}

// Test HSVToRGB
func TestHSVToRGB(t *testing.T) {
	hsv := HSV{H: 9.13, S: 0.72, V: 1.0}
	expectedRGB := RGB{R: 255, G: 99, B: 71}

	rgb := HSVToRGB(hsv)
	if rgb != expectedRGB {
		t.Errorf("HSVToRGB failed: expected %+v, got %+v", expectedRGB, rgb)
	}
}

// Test RGBToCMYK
func TestRGBToCMYK(t *testing.T) {
	rgb := RGB{R: 255, G: 99, B: 71}
	expectedCMYK := CMYK{C: 0.0, M: 0.611764705882353, Y: 0.7215686274509804, K: 0.0}

	cmyk := RGBToCMYK(rgb)
	if !compareFloat(cmyk.C, expectedCMYK.C) || !compareFloat(cmyk.M, expectedCMYK.M) || !compareFloat(cmyk.Y, expectedCMYK.Y) || !compareFloat(cmyk.K, expectedCMYK.K) {
		t.Errorf("RGBToCMYK failed: expected %+v, got %+v", expectedCMYK, cmyk)
	}
}

// Test CMYKToRGB
func TestCMYKToRGB(t *testing.T) {
	cmyk := CMYK{C: 0.0, M: 0.611764705882353, Y: 0.7215686274509804, K: 0.0}
	expectedRGB := RGB{R: 255, G: 99, B: 71}

	rgb := CMYKToRGB(cmyk)
	if rgb != expectedRGB {
		t.Errorf("CMYKToRGB failed: expected %+v, got %+v", expectedRGB, rgb)
	}
}

// Utility function to compare floating-point numbers
func compareFloat(a, b float64) bool {
	const epsilon = 1e-5
	return math.Abs(a-b) < epsilon
}
