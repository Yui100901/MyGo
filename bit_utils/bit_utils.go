package bit_utils

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

// SetBit 设置某一位为1
func (b *BitArray) SetBit(index int) error {
	if index < 0 || index >= b.bitLen {
		return errors.New("index out of bounds")
	}
	b.setBit(index)
	return nil
}

// 内部方法set
func (b *BitArray) setBit(index int) {
	b.data[index>>3] |= 1 << uint(7-(index&7))
}

// SetByte 设置某个字节
func (b *BitArray) SetByte(index int, byte2 byte) error {
	if index < 0 || index >= b.ByteLen() {
		return errors.New("index out of bounds")
	}
	b.setByte(index, byte2)
	return nil
}

func (b *BitArray) setByte(index int, byte2 byte) {
	b.data[index] = byte2
}

// ClearBit 清除某一位
func (b *BitArray) ClearBit(index int) error {
	if index < 0 || index >= b.bitLen {
		return errors.New("index out of bounds")
	}
	b.clearBit(index)
	return nil
}

// 内部方法clear
func (b *BitArray) clearBit(index int) {
	b.data[index>>3] &^= 1 << uint(7-(index&7))
}

// ClearByte 清除某个字节
func (b *BitArray) ClearByte(index int) error {
	if index < 0 || index >= b.ByteLen() {
		return errors.New("index out of bounds")
	}
	b.clearByte(index)
	return nil
}

func (b *BitArray) clearByte(index int) {
	b.data[index] = 0x00
}

// GetBit 取某一位
func (b *BitArray) GetBit(index int) (byte, error) {
	if index < 0 || index >= b.bitLen {
		return 0, errors.New("index out of bounds")
	}
	return b.getBit(index), nil
}

func (b *BitArray) getBit(index int) byte {
	return (b.data[index>>3] >> uint(7-(index&7))) & 1
}

// GetByte 取某一字节
func (b *BitArray) GetByte(byteIndex int) (byte, error) {
	if byteIndex < 0 || byteIndex >= b.ByteLen() {
		return 0, errors.New("index out of bounds")
	}
	return b.getByte(byteIndex), nil
}

func (b *BitArray) getByte(byteIndex int) byte {
	return b.data[byteIndex]
}

// IsOne 判断位是否为1
func (b *BitArray) IsOne(index int) (bool, error) {
	v, err := b.GetBit(index)
	return v == 1, err
}

// IsZero 判断位是否为0
func (b *BitArray) IsZero(index int) (bool, error) {
	v, err := b.GetBit(index)
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

// SetAll 设置所有位为1 - 修复末尾处理
func (b *BitArray) SetAll() {
	for i := range b.data {
		b.data[i] = 0xFF
	}
	// 修复末尾余位处理
	if remainder := b.bitLen & 7; remainder > 0 && len(b.data) > 0 {
		lastByteIndex := len(b.data) - 1
		// 清除超出位数的位
		mask := byte(0xFF >> (8 - remainder))
		b.data[lastByteIndex] &= mask
		b.data[lastByteIndex] |= mask
	}
}

// Count 计算设置为1的位数 - 改进实现
func (b *BitArray) Count() int {
	count := 0
	fullBytes := b.bitLen >> 3

	// 计算完整字节的1位数
	for i := 0; i < fullBytes; i++ {
		byteVal := b.data[i]
		// Brian Kernighan算法计算字节中1的个数
		for byteVal != 0 {
			count++
			byteVal &= byteVal - 1
		}
	}

	// 处理不完整的最后一个字节
	if remainder := b.bitLen & 7; remainder > 0 && len(b.data) > 0 {
		lastByte := b.data[len(b.data)-1]
		// 只计算有效位
		mask := byte(0xFF >> (8 - remainder))
		lastByte &= mask
		for lastByte != 0 {
			count++
			lastByte &= lastByte - 1
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
		if v, _ := b.GetBit(i); v == 1 {
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

// SetRange 设置指定范围的位为1
func (b *BitArray) SetRange(start, length int) error {
	if start < 0 || length < 0 || start+length > b.bitLen {
		return errors.New("invalid range")
	}

	// 优化：批量设置字节级别的操作
	end := start + length
	startByte := start >> 3
	endByte := (end - 1) >> 3

	if startByte == endByte {
		// 在同一个字节内
		for i := start; i < end; i++ {
			b.setBit(i)
		}
	} else {
		// 跨越多个字节
		// 设置第一个不完整字节
		firstByteEnd := (startByte + 1) << 3
		for i := start; i < firstByteEnd && i < end; i++ {
			b.setBit(i)
		}

		// 设置中间的完整字节
		for byteIdx := startByte + 1; byteIdx < endByte; byteIdx++ {
			b.data[byteIdx] = 0xFF
		}

		// 设置最后一个不完整字节
		lastByteStart := endByte << 3
		for i := lastByteStart; i < end; i++ {
			b.setBit(i)
		}
	}

	return nil
}

// ClearRange 清除指定范围的位
func (b *BitArray) ClearRange(start, length int) error {
	if start < 0 || length < 0 || start+length > b.bitLen {
		return errors.New("invalid range")
	}

	// 优化：批量清除字节级别的操作
	end := start + length
	startByte := start >> 3
	endByte := (end - 1) >> 3

	if startByte == endByte {
		// 在同一个字节内
		for i := start; i < end; i++ {
			b.clearBit(i)
		}
	} else {
		// 跨越多个字节
		// 清除第一个不完整字节
		firstByteEnd := (startByte + 1) << 3
		for i := start; i < firstByteEnd && i < end; i++ {
			b.clearBit(i)
		}

		// 清除中间的完整字节
		for byteIdx := startByte + 1; byteIdx < endByte; byteIdx++ {
			b.data[byteIdx] = 0x00
		}

		// 清除最后一个不完整字节
		lastByteStart := endByte << 3
		for i := lastByteStart; i < end; i++ {
			b.clearBit(i)
		}
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

// Not 执行按位取反操作
func (b *BitArray) Not() *BitArray {
	result, _ := NewBitArray(b.bitLen)
	for i := 0; i < len(b.data); i++ {
		result.data[i] = ^b.data[i]
	}

	// 清除超出位数的位
	if remainder := b.bitLen & 7; remainder > 0 && len(result.data) > 0 {
		lastByteIndex := len(result.data) - 1
		mask := byte(0xFF >> (8 - remainder))
		result.data[lastByteIndex] &= mask
	}

	return result
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

	// 字节级别复制
	fullBytes := b.bitLen >> 3
	copy(result.data[:fullBytes], b.data[:fullBytes])

	// 处理第一个数组的剩余位
	if remainder := b.bitLen & 7; remainder > 0 {
		for i := fullBytes << 3; i < b.bitLen; i++ {
			if v, _ := b.GetBit(i); v == 1 {
				result.setBit(i)
			}
		}
	}

	// 复制第二个位数组
	for i := 0; i < other.bitLen; i++ {
		if v, _ := other.GetBit(i); v == 1 {
			result.setBit(b.bitLen + i)
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
			if v, _ := arr.GetBit(i); v == 1 {
				result.setBit(currentPos + i)
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
		if v, _ := b.GetBit(start + i); v == 1 {
			result.setBit(i)
		}
	}

	return result, nil
}

// SliceByte 返回字节的子切片 [start:end)
func (b *BitArray) SliceByte(start, end int) ([]byte, error) {
	if start < 0 || end < start || end > b.ByteLen() {
		return nil, errors.New("invalid slice range")
	}
	result := make([]byte, end-start)
	copy(result, b.data[start:end])
	return result, nil
}

// FindFirst 查找第一个设置为1的位的索引，未找到返回-1
func (b *BitArray) FindFirst() int {
	for i := 0; i < b.bitLen; i++ {
		if v, _ := b.GetBit(i); v == 1 {
			return i
		}
	}
	return -1
}

// FindLast 查找最后一个设置为1的位的索引，未找到返回-1
func (b *BitArray) FindLast() int {
	for i := b.bitLen - 1; i >= 0; i-- {
		if v, _ := b.GetBit(i); v == 1 {
			return i
		}
	}
	return -1
}

// IsEmpty 判断是否所有位都为0
func (b *BitArray) IsEmpty() bool {
	return b.Count() == 0
}

// IsFull 判断是否所有位都为1
func (b *BitArray) IsFull() bool {
	return b.Count() == b.bitLen
}
