package db

//
// @Author yfy2001
// @Date 2025/8/18 10 21
//

type Page[T any] struct {
	Records  []*T  `json:"records"`  // 当前页数据
	Total    int64 `json:"total"`    // 总记录数
	Current  int   `json:"current"`  // 当前页码
	PageSize int   `json:"pageSize"` // 每页大小

	HasPrev    bool  `json:"hasPrev"`    // 是否还有上一页
	HasNext    bool  `json:"hasNext"`    // 是否还有下一页
	TotalPages int64 `json:"totalPages"` // 总页数
}

func Paginate[T any](records []*T, total int64, current int, pageSize int) *Page[T] {
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	if totalPages == 0 {
		totalPages = 1
	}

	return &Page[T]{
		Records:    records,
		Total:      total,
		Current:    current,
		PageSize:   pageSize,
		HasPrev:    current > 1,
		HasNext:    int64(current) < totalPages,
		TotalPages: totalPages,
	}
}

// IsEmpty 判断是否为空页
func (p *Page[T]) IsEmpty() bool {
	return len(p.Records) == 0
}

// IsFirst 判断是否为第一页
func (p *Page[T]) IsFirst() bool {
	return p.Current == 1
}

// IsLast 判断是否为最后一页
func (p *Page[T]) IsLast() bool {
	return int64(p.Current) >= p.TotalPages
}
