package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/dictdata"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/dicttype"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
)

var _ biz.DictDataRepo = (*dictDataRepo)(nil)

type dictDataRepo struct {
	data *Data
	log  *log.Helper
}

func NewDictDataRepo(data *Data, logger log.Logger) biz.DictDataRepo {
	return &dictDataRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *dictDataRepo) List(ctx context.Context, params *biz.ListDictDataRequest) (*biz.ListDictDataResponse, error) {
	query := r.data.db.DictData.Query().
		Order(ent.Desc(dicttype.FieldCreatedAt))

	if v := params.DictTypeIds; len(v) > 0 {
		query = query.Where(dictdata.DictTypeIDIn(v...))
	}

	res, err := enthelper.Pagination[*ent.DictData, *ent.DictDataQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.DictData, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.DictData{
			ID:          v.ID,
			DictTypeId:  v.DictTypeID,
			Label:       v.Label,
			Value:       v.Value,
			Weight:      v.Weight,
			Status:      string(v.Status),
			Extension:   v.Extension,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
			DeletedAt:   v.DeletedAt,
		})
	}

	return &biz.ListDictDataResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}
