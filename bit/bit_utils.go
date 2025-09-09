package bit

import "errors"

// BitArray 表示一个位数组，使用紧凑存储
type BitArray struct {
	bitLen int
	data   []byte
}

// NewBitArray 创建一个指定长度的位数组，初始值全为 0
func NewBitArray(length int) (*BitArray, error) {
	if length < 0 {
		return nil, errors.New("length must be non-negative")
	}
	if length == 0 {
		return &BitArray{bitLen: 0, data: nil}, nil
	}

	byteLength := (length + 7) >> 3
	return &BitArray{
		bitLen: length,
		data:   make([]byte, byteLength),
	}, nil
}

// NewBitArrayFromExtract 从字节数组中提取位数组
func NewBitArrayFromExtract(data []byte, startIndex, length int) (*BitArray, error) {
	if startIndex < 0 || length < 0 {
		return nil, errors.New("invalid index or length")
	}
	if length == 0 {
		return NewBitArray(0)
	}

	totalBits := len(data) << 3
	if startIndex >= totalBits {
		return nil, errors.New("index out of bounds")
	}
	if startIndex+length > totalBits {
		return nil, errors.New("extraction range out of bounds")
	}

	bitArray, _ := NewBitArray(length)

	byteIndex := startIndex >> 3
	bitOffset := uint(startIndex & 7)

	for i := 0; i < length; i++ {
		if bitOffset == 8 {
			byteIndex++
			bitOffset = 0
		}

		if (data[byteIndex]>>(7-bitOffset))&1 == 1 {
			bitArray.data[i>>3] |= 1 << (7 - uint(i&7))
		}

		bitOffset++
	}

	return bitArray, nil
}

// NewBitArrayFromBytes 将字节切片转换成位数组
func NewBitArrayFromBytes(data []byte) *BitArray {
	if len(data) == 0 {
		return &BitArray{bitLen: 0, data: nil}
	}
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return &BitArray{
		bitLen: len(data) << 3,
		data:   dataCopy,
	}
}

func (b *BitArray) BitLen() int  { return b.bitLen }
func (b *BitArray) ByteLen() int { return len(b.data) }

func (b *BitArray) Set(index int) error {
	if index < 0 || index >= b.bitLen {
		return errors.New("index out of bounds")
	}
	b.set(index)
	return nil
}

// 内部方法set
func (b *BitArray) set(index int) {
	b.data[index>>3] |= 1 << uint(7-(index&7))
}

func (b *BitArray) Clear(index int) error {
	if index < 0 || index >= b.bitLen {
		return errors.New("index out of bounds")
	}
	b.data[index>>3] &^= 1 << uint(7-(index&7))
	return nil
}

// 内部方法set
func (b *BitArray) clear(index int) {
	b.data[index>>3] &^= 1 << uint(7-(index&7))
}

func (b *BitArray) Get(index int) (byte, error) {
	if index < 0 || index >= b.bitLen {
		return 0, errors.New("index out of bounds")
	}
	return b.get(index), nil
}

func (b *BitArray) get(index int) byte {
	return (b.data[index>>3] >> uint(7-(index&7))) & 1
}

func (b *BitArray) IsOne(index int) (bool, error) {
	v, err := b.Get(index)
	return v == 1, err
}

func (b *BitArray) IsZero(index int) (bool, error) {
	v, err := b.Get(index)
	return v == 0, err
}

func (b *BitArray) Toggle(index int) error {
	if index < 0 || index >= b.bitLen {
		return errors.New("index out of bounds")
	}
	b.data[index>>3] ^= 1 << uint(7-(index&7))
	return nil
}

func (b *BitArray) ClearAll() {
	for i := range b.data {
		b.data[i] = 0
	}
}

func (b *BitArray) SetAll() {
	for i := range b.data {
		b.data[i] = 0xFF
	}
	if remainder := b.bitLen & 7; remainder > 0 {
		lastByteIndex := len(b.data) - 1
		b.data[lastByteIndex] &= 0xFF << (8 - remainder)
	}
}

func (b *BitArray) Count() int {
	count := 0
	for _, byteVal := range b.data {
		for byteVal != 0 {
			count++
			byteVal &= byteVal - 1
		}
	}
	if remainder := b.bitLen & 7; remainder > 0 && len(b.data) > 0 {
		lastByte := b.data[len(b.data)-1]
		extraBits := lastByte & (0xFF >> remainder)
		for extraBits != 0 {
			count--
			extraBits &= extraBits - 1
		}
	}
	return count
}

func (b *BitArray) ToBytes() []byte {
	if len(b.data) == 0 {
		return nil
	}
	result := make([]byte, len(b.data))
	copy(result, b.data)
	return result
}

func (b *BitArray) String() string {
	if b.bitLen == 0 {
		return ""
	}
	result := make([]byte, b.bitLen)
	for i := 0; i < b.bitLen; i++ {
		if v, _ := b.Get(i); v == 1 {
			result[i] = '1'
		} else {
			result[i] = '0'
		}
	}
	return string(result)
}

func (b *BitArray) Clone() *BitArray {
	if len(b.data) == 0 {
		return &BitArray{bitLen: b.bitLen, data: nil}
	}
	clone := &BitArray{
		bitLen: b.bitLen,
		data:   make([]byte, len(b.data)),
	}
	copy(clone.data, b.data)
	return clone
}

func (b *BitArray) Equals(other *BitArray) bool {
	if other == nil || b.bitLen != other.bitLen {
		return false
	}
	for i := 0; i < len(b.data); i++ {
		if b.data[i] != other.data[i] {
			return false
		}
	}
	return true
}

func (b *BitArray) SetRange(start, length int) error {
	if start < 0 || length < 0 || start+length > b.bitLen {
		return errors.New("invalid range")
	}
	for i := 0; i < length; i++ {
		b.set(start + i)
	}
	return nil
}

func (b *BitArray) ClearRange(start, length int) error {
	if start < 0 || length < 0 || start+length > b.bitLen {
		return errors.New("invalid range")
	}
	for i := 0; i < length; i++ {
		b.clear(start + i)
	}
	return nil
}

func (b *BitArray) And(other *BitArray) (*BitArray, error) {
	if b.bitLen != other.bitLen {
		return nil, errors.New("bit arrays must have same length")
	}
	result, _ := NewBitArray(b.bitLen)
	for i := 0; i < len(b.data); i++ {
		result.data[i] = b.data[i] & other.data[i]
	}
	return result, nil
}

func (b *BitArray) Or(other *BitArray) (*BitArray, error) {
	if b.bitLen != other.bitLen {
		return nil, errors.New("bit arrays must have same length")
	}
	result, _ := NewBitArray(b.bitLen)
	for i := 0; i < len(b.data); i++ {
		result.data[i] = b.data[i] | other.data[i]
	}
	return result, nil
}

// Xor 执行按位异或操作，返回新的位数组
func (b *BitArray) Xor(other *BitArray) (*BitArray, error) {
	if b.bitLen != other.bitLen {
		return nil, errors.New("bit arrays must have same length")
	}
	result, _ := NewBitArray(b.bitLen)
	for i := 0; i < len(b.data); i++ {
		result.data[i] = b.data[i] ^ other.data[i]
	}
	return result, nil
}

// Append 将另一个位数组追加到当前位数组末尾
func (b *BitArray) Append(other *BitArray) (*BitArray, error) {
	if other == nil || other.bitLen == 0 {
		return b.Clone(), nil
	}
	if b.bitLen == 0 {
		return other.Clone(), nil
	}

	totalLength := b.bitLen + other.bitLen
	result, _ := NewBitArray(totalLength)

	// 复制第一个位数组
	for i := 0; i < b.bitLen; i++ {
		if v, _ := b.Get(i); v == 1 {
			result.set(i)
		}
	}

	// 复制第二个位数组
	for i := 0; i < other.bitLen; i++ {
		if v, _ := other.Get(i); v == 1 {
			result.set(b.bitLen + i)
		}
	}

	return result, nil
}

// Concat 将多个位数组拼接成一个新的位数组
func Concat(arrays ...*BitArray) (*BitArray, error) {
	if len(arrays) == 0 {
		return NewBitArray(0)
	}

	totalLength := 0
	for _, arr := range arrays {
		if arr != nil {
			totalLength += arr.bitLen
		}
	}

	if totalLength == 0 {
		return NewBitArray(0)
	}

	result, _ := NewBitArray(totalLength)
	currentPos := 0

	for _, arr := range arrays {
		if arr == nil || arr.bitLen == 0 {
			continue
		}
		for i := 0; i < arr.bitLen; i++ {
			if v, _ := arr.Get(i); v == 1 {
				result.set(currentPos + i)
			}
		}
		currentPos += arr.bitLen
	}

	return result, nil
}

// Slice 返回位数组的子切片 [start:end)
func (b *BitArray) Slice(start, end int) (*BitArray, error) {
	if start < 0 || end < start || end > b.bitLen {
		return nil, errors.New("invalid slice range")
	}

	length := end - start
	if length == 0 {
		return NewBitArray(0)
	}

	result, _ := NewBitArray(length)
	for i := 0; i < length; i++ {
		if v, _ := b.Get(start + i); v == 1 {
			result.set(i)
		}
	}

	return result, nil
}
