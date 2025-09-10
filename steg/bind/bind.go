package main

import (
	"bytes"
	"errors"
	"io"
)

type Bind struct {
	carrierOffset int64
}

// Embed 返回一个组合的 Reader，如果无法 Seek 则一次性读入 carrier
func (b *Bind) Embed(carrier io.Reader, payload io.Reader) (io.Reader, error) {
	if rs, ok := carrier.(io.ReadSeeker); ok {
		// 支持 Seek，直接获取长度
		cur, _ := rs.Seek(0, io.SeekCurrent)
		end, err := rs.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		b.carrierOffset = end
		_, _ = rs.Seek(cur, io.SeekStart)
		return io.MultiReader(rs, payload), nil
	}

	// 不支持 Seek，读入内存
	carrierData, err := io.ReadAll(carrier)
	if err != nil {
		return nil, err
	}
	b.carrierOffset = int64(len(carrierData))

	return io.MultiReader(bytes.NewReader(carrierData), payload), nil
}

// Extract 如果支持 Seek 则直接跳转，否则一次性读入并截取
func (b *Bind) Extract(carrier io.Reader) (io.Reader, error) {
	if b.carrierOffset < 0 {
		return nil, errors.New("未知 carrierOffset，无法提取")
	}

	if rs, ok := carrier.(io.ReadSeeker); ok {
		_, err := rs.Seek(b.carrierOffset, io.SeekStart)
		if err != nil {
			return nil, err
		}
		return rs, nil
	}

	// 不支持 Seek，读入内存后截取
	allData, err := io.ReadAll(carrier)
	if err != nil {
		return nil, err
	}
	if b.carrierOffset > int64(len(allData)) {
		return nil, errors.New("carrierOffset 超出数据范围")
	}
	return bytes.NewReader(allData[b.carrierOffset:]), nil
}
