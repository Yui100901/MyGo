package main

import (
	"errors"
	"io"
)

type Bind struct {
	carrierFormat string
	carrierOffset int64 // 记录载体文件的长度
}

func NewBind(format string) *Bind {
	return &Bind{
		carrierFormat: format,
	}
}

// Embed Embed: 返回一个组合的 Reader，不会一次性读入内存
func (b *Bind) Embed(carrier io.Reader, payload io.Reader) (io.Reader, error) {
	// 如果 carrier 是 io.Seeker，可以获取长度
	if seeker, ok := carrier.(io.Seeker); ok {
		// 记录当前位置
		cur, _ := seeker.Seek(0, io.SeekCurrent)
		end, err := seeker.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		b.carrierOffset = end
		// 回到原位置
		_, _ = seeker.Seek(cur, io.SeekStart)
	} else {
		// 如果不是 Seeker，就无法提前知道长度
		b.carrierOffset = -1
	}

	// MultiReader 会先读 carrier，再读 payload
	return io.MultiReader(carrier, payload), nil
}

// Extract Extract: 从 carrier 中跳过前面的 carrierOffset 字节，返回 payload 部分
func (b *Bind) Extract(carrier io.Reader) (io.Reader, error) {
	if b.carrierOffset < 0 {
		return nil, errors.New("未知 carrierOffset，无法提取")
	}

	seeker, ok := carrier.(io.ReadSeeker)
	if !ok {
		return nil, errors.New("carrier 不支持 Seek，无法流式提取")
	}

	_, err := seeker.Seek(b.carrierOffset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// 返回从偏移位置开始的 Reader
	return seeker, nil
}
