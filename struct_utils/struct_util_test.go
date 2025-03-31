package struct_utils

import "testing"

//
// @Author yfy2001
// @Date 2025/3/31 15 09
//

// 基本结构体（用于基本类型字段测试）
type Basic struct {
	A int
	B string
}

// 源结构体包含目标结构体没有的额外字段
type BasicExtra struct {
	A int
	B string
	C bool // 额外字段，不会复制
}

// 嵌套结构体测试
type Nested struct {
	X int
	Y string
}

type WithNested struct {
	N Nested
}

// 指针字段测试（既包含基本类型指针，也包含结构体指针）
type PointerTest struct {
	PtrInt    *int
	PtrStr    *string
	PtrNested *Nested
}

// 字段类型不匹配，测试当字段类型不同时不进行赋值
type Mismatch struct {
	A int
}

type MismatchDst struct {
	A string // 类型不同，不会复制
}

// 内嵌结构体测试（匿名字段）
type EmbeddedSrc struct {
	Basic  // 内嵌 Basic 结构体
	Extra  string
	hidden int // 未导出字段，不会被复制
}

type EmbeddedDst struct {
	Basic
	Extra  string
	hidden int
}

// 单元测试：逐个测试，同时在转换前后打印出源结构体和目标结构体
func TestConvertStruct(t *testing.T) {
	// Test 1: 基本字段直接赋值测试
	t.Run("Basic conversion", func(t *testing.T) {
		src := Basic{A: 10, B: "hello"}
		dst := Basic{}
		t.Logf("Before conversion: src: %+v, dst: %+v", src, dst)
		if err := ConvertStruct(&src, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("After conversion: src: %+v, dst: %+v", src, dst)
		if dst.A != 10 || dst.B != "hello" {
			t.Errorf("expected {10, hello}, got %+v", dst)
		}
	})

	// Test 2: 源结构体包含目标不存在的额外字段
	t.Run("Extra field in source", func(t *testing.T) {
		src := BasicExtra{A: 20, B: "world", C: true}
		dst := Basic{}
		t.Logf("Before conversion: src: %+v, dst: %+v", src, dst)
		if err := ConvertStruct(&src, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("After conversion: src: %+v, dst: %+v", src, dst)
		if dst.A != 20 || dst.B != "world" {
			t.Errorf("expected {20, world}, got %+v", dst)
		}
	})

	// Test 3: 嵌套结构体转换，递归调用测试
	t.Run("Nested struct conversion", func(t *testing.T) {
		src := WithNested{N: Nested{X: 100, Y: "nested"}}
		dst := WithNested{}
		t.Logf("Before conversion: src: %+v, dst: %+v", src, dst)
		if err := ConvertStruct(&src, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("After conversion: src: %+v, dst: %+v", src, dst)
		if dst.N.X != 100 || dst.N.Y != "nested" {
			t.Errorf("expected {100, nested}, got %+v", dst.N)
		}
	})

	// Test 4: 指针字段转换（非 nil 指针）测试
	t.Run("Pointer fields conversion non-nil", func(t *testing.T) {
		a := 42
		str := "pointer test"
		src := PointerTest{
			PtrInt:    &a,
			PtrStr:    &str,
			PtrNested: &Nested{X: 500, Y: "ptr nested"},
		}
		dst := PointerTest{}
		t.Logf("Before conversion: src: %+v, dst: %+v", src, dst)
		if err := ConvertStruct(&src, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("After conversion: src: %+v, dst: %+v", src, dst)
		if dst.PtrInt == nil || *dst.PtrInt != 42 {
			t.Errorf("expected PtrInt to be 42, got %v", dst.PtrInt)
		}
		if dst.PtrStr == nil || *dst.PtrStr != "pointer test" {
			t.Errorf("expected PtrStr to be 'pointer test', got %v", dst.PtrStr)
		}
		if dst.PtrNested == nil || dst.PtrNested.X != 500 || dst.PtrNested.Y != "ptr nested" {
			t.Errorf("expected PtrNested to be {500, ptr nested}, got %+v", dst.PtrNested)
		}
	})

	// Test 5: 指针字段转换，当源字段为 nil 时目标应被置为零值
	t.Run("Pointer fields conversion with nils", func(t *testing.T) {
		src := PointerTest{
			PtrInt:    nil,
			PtrStr:    nil,
			PtrNested: nil,
		}
		// 初始化目标结构体含非 nil 的指针字段
		dst := PointerTest{
			PtrInt:    new(int),
			PtrStr:    new(string),
			PtrNested: &Nested{X: 1, Y: "initial"},
		}
		t.Logf("Before conversion: src: %+v, dst: %+v", src, dst)
		if err := ConvertStruct(&src, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("After conversion: src: %+v, dst: %+v", src, dst)
		if dst.PtrInt != nil && *dst.PtrInt != 0 {
			t.Errorf("expected PtrInt to be zero, got %v", *dst.PtrInt)
		}
		if dst.PtrStr != nil && *dst.PtrStr != "" {
			t.Errorf("expected PtrStr to be empty, got %v", *dst.PtrStr)
		}
		if dst.PtrNested != nil {
			t.Errorf("expected PtrNested to be nil, got %+v", dst.PtrNested)
		}
	})

	// Test 6: 字段类型不匹配时，不复制，应保持目标字段原始值
	t.Run("Mismatched field types", func(t *testing.T) {
		src := Mismatch{A: 99}
		dst := MismatchDst{A: "should remain unchanged"}
		t.Logf("Before conversion: src: %+v, dst: %+v", src, dst)
		if err := ConvertStruct(&src, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("After conversion: src: %+v, dst: %+v", src, dst)
		if dst.A != "should remain unchanged" {
			t.Errorf("expected A to remain unchanged, got %v", dst.A)
		}
	})

	// Test 7: 内嵌结构体转换测试
	t.Run("Embedded struct conversion", func(t *testing.T) {
		src := EmbeddedSrc{
			Basic:  Basic{A: 55, B: "embedded"},
			Extra:  "extra info",
			hidden: 999, // 未导出字段，不会复制
		}
		dst := EmbeddedDst{
			Basic:  Basic{A: 0, B: ""},
			Extra:  "",
			hidden: 0,
		}
		t.Logf("Before conversion: src: %+v, dst: %+v", src, dst)
		if err := ConvertStruct(&src, &dst); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		t.Logf("After conversion: src: %+v, dst: %+v", src, dst)
		if dst.Basic.A != 55 || dst.Basic.B != "embedded" {
			t.Errorf("expected Basic fields to be copied, got %+v", dst.Basic)
		}
		if dst.Extra != "extra info" {
			t.Errorf("expected Extra to be 'extra info', got %s", dst.Extra)
		}
		// unexported 字段 hidden 不会被复制，应保持目标值
		if dst.hidden != 0 {
			t.Errorf("expected hidden field to remain 0, got %d", dst.hidden)
		}
	})

	// Test 8: 参数非指针应返回错误
	t.Run("Non-pointer arguments", func(t *testing.T) {
		src := Basic{A: 1, B: "non-pointer"}
		dst := Basic{}
		t.Logf("Before conversion (non-pointer src): src: %+v, dst: %+v", src, dst)
		err := ConvertStruct(src, &dst)
		if err == nil {
			t.Errorf("expected error for non-pointer src, got nil")
		}
		t.Logf("After conversion (non-pointer src): src: %+v, dst: %+v", src, dst)
		t.Logf("Before conversion (non-pointer dst): src: %+v, dst: %+v", src, dst)
		err = ConvertStruct(&src, dst)
		if err == nil {
			t.Errorf("expected error for non-pointer dst, got nil")
		}
		t.Logf("After conversion (non-pointer dst): src: %+v, dst: %+v", src, dst)
	})

	// Test 9: 指针却不指向结构体，应返回错误
	t.Run("Pointers not to structs", func(t *testing.T) {
		i := 100
		j := 200
		t.Logf("Before conversion: src: %d, dst: %d", i, j)
		err := ConvertStruct(&i, &j)
		if err == nil {
			t.Errorf("expected error for pointers not to structs, got nil")
		}
		t.Logf("After conversion: src: %d, dst: %d", i, j)
	})
}
