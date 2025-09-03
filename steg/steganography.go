package steg

import (
	"io"
)

//
// @Author yfy2001
// @Date 2025/8/26 17 26
//

// Steganography 隐写接口
// 负责将载荷信息写入到载体
type Steganography interface {
	Embed(carrier io.Reader, payload io.Reader) (io.Reader, error)
	Extract(carrier io.Reader) (io.Reader, error)
}
