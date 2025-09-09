package bit

import "errors"

// BitArray 表示一个位数组，使用紧凑存储
type BitArray struct {
	bitLen int    // 位长度
	data   []byte // 每个字节存储8个位，提高内存效率
}

// NewBitArray 创建一个指定长度的位数组，初始值全为 0
func NewBitArray(length int) *BitArray {
	if length < 0 {
		panic("length must be non-negative")
	}
	if length == 0 {
		return &BitArray{bitLen: 0, data: nil}
	}

	byteLength := (length + 7) >> 3
	return &BitArray{
		bitLen: length,
		data:   make([]byte, byteLength),
	}
}

// NewBitArrayFromExtract 从字节数组中提取位数组
func NewBitArrayFromExtract(data []byte, startIndex, length int) (*BitArray, error) {
	if startIndex < 0 || length < 0 {
		return nil, errors.New("invalid index or length")
	}
	if length == 0 {
		return NewBitArray(0), nil
	}

	totalBits := len(data) << 3
	if startIndex >= totalBits {
		return nil, errors.New("index out of bounds")
	}
	if startIndex+length > totalBits {
		return nil, errors.New("extraction range out of bounds")
	}

	bitArray := NewBitArray(length)

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

	// 创建副本，避免外部修改影响内部状态
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	return &BitArray{
		bitLen: len(data) << 3,
		data:   dataCopy,
	}
}

// BitLen 返回位长度
func (b *BitArray) BitLen() int {
	return b.bitLen
}

// ByteLen 返回字节长度
func (b *BitArray) ByteLen() int {
	return len(b.data)
}

// Set 将指定位置的位设置为 0 或 1，只接受 0 或 1
func (b *BitArray) Set(index int, value byte) {
	if value != 0 && value != 1 {
		panic("value must be 0 or 1")
	}

	if index < 0 || index >= b.bitLen {
		panic("index out of bounds")
	}

	byteIndex := index >> 3
	bitOffset := uint(7 - (index & 7))

	if value == 1 {
		b.data[byteIndex] |= 1 << bitOffset
	} else {
		b.data[byteIndex] &^= 1 << bitOffset
	}
}

// SetBit 将指定位置的位设置为 1
func (b *BitArray) SetBit(index int) {
	if index < 0 || index >= b.bitLen {
		panic("index out of bounds")
	}

	byteIndex := index >> 3
	// 修正：统一位偏移计算方式
	b.data[byteIndex] |= 1 << uint(7-(index&7))
}

// ClearBit 将指定位置的位设置为 0
func (b *BitArray) ClearBit(index int) {
	if index < 0 || index >= b.bitLen {
		panic("index out of bounds")
	}

	byteIndex := index >> 3
	// 修正：统一位偏移计算方式
	b.data[byteIndex] &^= 1 << uint(7-(index&7))
}

// Get 获取指定位置的位值
func (b *BitArray) Get(index int) byte {
	if index < 0 || index >= b.bitLen {
		panic("index out of bounds")
	}

	byteIndex := index >> 3
	// 修正：统一位偏移计算方式
	return (b.data[byteIndex] >> uint(7-(index&7))) & 1
}

// IsOne 返回指定位置的位是否为 1
func (b *BitArray) IsOne(index int) bool {
	return b.Get(index) == 1
}

// IsZero 返回指定位置的位是否为 0
func (b *BitArray) IsZero(index int) bool {
	return b.Get(index) == 0
}

// Toggle 翻转指定位置的位
func (b *BitArray) Toggle(index int) {
	if index < 0 || index >= b.bitLen {
		panic("index out of bounds")
	}

	byteIndex := index >> 3
	// 修正：统一位偏移计算方式
	b.data[byteIndex] ^= 1 << uint(7-(index&7))
}

// Clear 将所有位清零
func (b *BitArray) Clear() {
	for i := range b.data {
		b.data[i] = 0
	}
}

// SetAll 将所有位设置为 1
func (b *BitArray) SetAll() {
	for i := range b.data {
		b.data[i] = 0xFF
	}

	// 清除最后一个字节中的多余位
	if remainder := b.bitLen & 7; remainder > 0 {
		lastByteIndex := len(b.data) - 1
		b.data[lastByteIndex] &= 0xFF << (8 - remainder)
	}
}

// Count 返回设置为 1 的位的数量
func (b *BitArray) Count() int {
	count := 0
	for _, byteVal := range b.data {
		// 使用 Brian Kernighan 算法计算位数
		for byteVal != 0 {
			count++
			byteVal &= byteVal - 1
		}
	}

	// 修正：需要处理最后一个字节的多余位
	if remainder := b.bitLen & 7; remainder > 0 && len(b.data) > 0 {
		lastByte := b.data[len(b.data)-1]
		// 计算多余位中1的个数并减去
		extraBits := lastByte & (0xFF >> remainder)
		for extraBits != 0 {
			count--
			extraBits &= extraBits - 1
		}
	}

	return count
}

// ToBytes 返回底层字节数组的副本
func (b *BitArray) ToBytes() []byte {
	if len(b.data) == 0 {
		return nil
	}
	result := make([]byte, len(b.data))
	copy(result, b.data)
	return result
}

// String 返回位数组的字符串表示
func (b *BitArray) String() string {
	if b.bitLen == 0 {
		return ""
	}

	result := make([]byte, b.bitLen)
	for i := 0; i < b.bitLen; i++ {
		if b.Get(i) == 1 {
			result[i] = '1'
		} else {
			result[i] = '0'
		}
	}
	return string(result)
}

// Clone 创建位数组的深拷贝
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

// Equals 比较两个位数组是否相等
func (b *BitArray) Equals(other *BitArray) bool {
	if other == nil || b.bitLen != other.bitLen {
		return false
	}

	// 逐字节比较
	for i := 0; i < len(b.data); i++ {
		if b.data[i] != other.data[i] {
			return false
		}
	}

	return true
}

// 新增一些实用方法

// SetRange 设置指定范围的位为 1
func (b *BitArray) SetRange(start, length int) {
	if start < 0 || length < 0 || start+length > b.bitLen {
		panic("invalid range")
	}

	for i := 0; i < length; i++ {
		b.SetBit(start + i)
	}
}

// ClearRange 设置指定范围的位为 0
func (b *BitArray) ClearRange(start, length int) {
	if start < 0 || length < 0 || start+length > b.bitLen {
		panic("invalid range")
	}

	for i := 0; i < length; i++ {
		b.ClearBit(start + i)
	}
}

// And 执行按位与操作，返回新的位数组
func (b *BitArray) And(other *BitArray) *BitArray {
	if b.bitLen != other.bitLen {
		panic("bit arrays must have same length")
	}

	result := NewBitArray(b.bitLen)
	for i := 0; i < len(b.data); i++ {
		result.data[i] = b.data[i] & other.data[i]
	}
	return result
}

// Or 执行按位或操作，返回新的位数组
func (b *BitArray) Or(other *BitArray) *BitArray {
	if b.bitLen != other.bitLen {
		panic("bit arrays must have same length")
	}

	result := NewBitArray(b.bitLen)
	for i := 0; i < len(b.data); i++ {
		result.data[i] = b.data[i] | other.data[i]
	}
	return result
}

// Xor 执行按位异或操作，返回新的位数组
func (b *BitArray) Xor(other *BitArray) *BitArray {
	if b.bitLen != other.bitLen {
		panic("bit arrays must have same length")
	}

	result := NewBitArray(b.bitLen)
	for i := 0; i < len(b.data); i++ {
		result.data[i] = b.data[i] ^ other.data[i]
	}
	return result
}

// Append 将另一个位数组追加到当前位数组末尾，返回新的位数组
func (b *BitArray) Append(other *BitArray) *BitArray {
	if other == nil || other.bitLen == 0 {
		return b.Clone()
	}
	if b.bitLen == 0 {
		return other.Clone()
	}

	totalLength := b.bitLen + other.bitLen
	result := NewBitArray(totalLength)

	// 复制第一个位数组
	for i := 0; i < b.bitLen; i++ {
		if b.Get(i) == 1 {
			result.SetBit(i)
		}
	}

	// 复制第二个位数组
	for i := 0; i < other.bitLen; i++ {
		if other.Get(i) == 1 {
			result.SetBit(b.bitLen + i)
		}
	}

	return result
}

// Concat 将多个位数组拼接成一个新的位数组
func Concat(arrays ...*BitArray) *BitArray {
	if len(arrays) == 0 {
		return NewBitArray(0)
	}

	// 计算总长度
	totalLength := 0
	for _, arr := range arrays {
		if arr != nil {
			totalLength += arr.bitLen
		}
	}

	if totalLength == 0 {
		return NewBitArray(0)
	}

	result := NewBitArray(totalLength)
	currentPos := 0

	// 依次复制每个数组
	for _, arr := range arrays {
		if arr == nil || arr.bitLen == 0 {
			continue
		}

		for i := 0; i < arr.bitLen; i++ {
			if arr.Get(i) == 1 {
				result.SetBit(currentPos + i)
			}
		}
		currentPos += arr.bitLen
	}

	return result
}

// Slice 返回位数组的子切片，类似切片操作 [start:end)
func (b *BitArray) Slice(start, end int) *BitArray {
	if start < 0 || end < start || end > b.bitLen {
		panic("invalid slice range")
	}

	length := end - start
	if length == 0 {
		return NewBitArray(0)
	}

	result := NewBitArray(length)
	for i := 0; i < length; i++ {
		if b.Get(start+i) == 1 {
			result.SetBit(i)
		}
	}

	return result
}
