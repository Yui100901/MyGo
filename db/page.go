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
