package bit

import "testing"

//
// @Author yfy2001
// @Date 2025/9/8 17 21
//

func TestExtractBitArray(t *testing.T) {
	data := []byte{0b10101010, 0b11001100}           // 0xAA, 0xCC
	bits, err := NewBitArrayFromExtract(data, 10, 3) // 从第4位开始取8个bit
	if err != nil {
		panic(err)
	}
	t.Log(bits)
}

func TestBinary(t *testing.T) {
	data := []byte{0b1010, 0b11}
	t.Log(len(data))
}
