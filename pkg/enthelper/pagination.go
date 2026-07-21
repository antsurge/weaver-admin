package enthelper

import (
	"context"
)

// 定义泛型接口约束
type PaginatableQuery[T any, Q any] interface {
	Count(ctx context.Context) (int, error)
	All(ctx context.Context) ([]T, error)
	Offset(n int) Q
	Limit(n int) Q
}

type PaginationResult[T any] struct {
	Total int `json:"total"`
	Data  []T `json:"data"`
}

type PaginationParams struct {
	CurrentPage int `form:"currentPage"`
	PageSize    int `form:"pageSize"`
}

func (p *PaginationParams) Normalize() {
	if p.CurrentPage <= 0 {
		p.CurrentPage = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 0
	}
	// 限制最大页数，防止数据库压力过大
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Pagination 分页函数
func Pagination[T any, Q PaginatableQuery[T, Q]](
	ctx context.Context,
	query Q,
	params PaginationParams,
) (*PaginationResult[T], error) {
	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &PaginationResult[T]{Total: 0, Data: []T{}}, nil
	}

	// pageSize=0 表示不分页
	if params.PageSize == 0 {
		list, err := query.All(ctx)
		if err != nil {
			return nil, err
		}
		return &PaginationResult[T]{Total: total, Data: list}, nil
	}

	// 计算分页
	if params.CurrentPage <= 0 {
		params.CurrentPage = 1
	}
	offset := (params.CurrentPage - 1) * params.PageSize

	list, err := query.Offset(offset).Limit(params.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	return &PaginationResult[T]{Total: total, Data: list}, nil
}
