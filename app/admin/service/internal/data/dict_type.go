package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/dicttype"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
)

var _ biz.DictTypeRepo = (*dictTypeRepo)(nil)

type dictTypeRepo struct {
	data *Data
	log  *log.Helper
}

func NewDictTypeRepo(data *Data, logger log.Logger) biz.DictTypeRepo {
	return &dictTypeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// 列表（支持分页）
func (r *dictTypeRepo) ListDictType(ctx context.Context, params *biz.ListDictTypeRequest) (*biz.ListDictTypeResponse, error) {
	query := r.data.db.DictType.Query().
		Order(ent.Desc(dicttype.FieldCreatedAt))

	res, err := enthelper.Pagination[*ent.DictType, *ent.DictTypeQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.DictType, 0, res.Total)
	// 转换为biz结构
	for _, v := range res.Data {
		data = append(data, &biz.DictType{
			ID:          v.ID,
			Name:        v.Name,
			Code:        v.Code,
			Status:      string(v.Status),
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
			DeletedAt:   v.DeletedAt,
		})
	}

	return &biz.ListDictTypeResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

// 创建
func (r *dictTypeRepo) CreateDictType(ctx context.Context, d *biz.DictType) error {
	return nil
}

// 编辑
func (r *dictTypeRepo) UpdateDictType(ctx context.Context, d *biz.DictType) error {
	return nil
}

// 删除
func (r *dictTypeRepo) DeleteDictType(ctx context.Context, ids []string) error {
	return nil
}
