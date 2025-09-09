package bit

//
// @Author yfy2001
// @Date 2025/9/9 20 26
//

import (
	"fmt"
	"testing"
)

func TestBit(t *testing.T) {
	x := 77
	fmt.Printf("x = %d\n", x)

	// 除以 2 的 n 次方：使用右移 n 位代替除法运算（x / 2^n）
	// 右移 3 位相当于除以 8（2^3）
	fmt.Printf("x / 8 = %d (优化为位移: x >> 3 = %d)\n", x/8, x>>3)

	// 乘以 2 的 n 次方：使用左移 n 位代替乘法运算（x * 2^n）
	// 左移 4 位相当于乘以 16（2^4）
	fmt.Printf("x * 16 = %d (优化为位移: x << 4 = %d)\n", x*16, x<<4)

	// 对 2 的 n 次方取模：使用位与运算代替取余（x % 2^n）
	// x & (2^n - 1) 等价于 x % 2^n，这里 15 = 2^4 - 1
	fmt.Printf("x %% 16 = %d (优化为位与: x & 15 = %d)\n", x%16, x&15)

	// 判断是否为 2 的幂：只有一个二进制位为 1 的数满足条件
	// 原理：2 的幂满足 x & (x - 1) == 0
	y := 64
	isPowerOfTwo := y != 0 && (y&(y-1)) == 0
	fmt.Printf("Is %d a power of two? %v\n", y, isPowerOfTwo)

	// 清除最低位的 1：使用 x & (x - 1) 快速将最低位的 1 变为 0
	// 适用于统计 1 的个数、位图压缩等
	z := 42 // 二进制：101010
	fmt.Printf("Before clearing lowest 1: %08b\n", z)
	z = z & (z - 1)
	fmt.Printf("After clearing lowest 1:  %08b\n", z)

	// 提取最低位的 1：使用 x & -x 快速定位最低位的 1
	// 适用于优先队列、调度器、位图定位等
	w := 42 // 二进制：101010
	lowestOne := w & -w
	fmt.Printf("Lowest 1 bit of %d: %08b\n", w, lowestOne)

	// 交换两个整数：使用异或操作，无需临时变量
	// 原理：a ^= b; b ^= a; a ^= b
	a, b := 5, 9
	a ^= b
	b ^= a
	a ^= b
	fmt.Printf("Swapped: a = %d, b = %d\n", a, b)

	// 判断奇偶性：使用 x & 1 判断最低位是否为 1
	// 如果最低位为 1，则为奇数；否则为偶数
	isOdd := x&1 == 1
	fmt.Printf("Is %d odd? %v\n", x, isOdd)
}
