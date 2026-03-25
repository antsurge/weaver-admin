package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/dictdata"
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
func (r *dictTypeRepo) ListDictType(ctx context.Context, params *biz.ListDictTypeRequest, opts ...*biz.ListDictTypeOption) (*biz.ListDictTypeResponse, error) {
	query := r.data.db.DictType.Query().
		Order(ent.Desc(dicttype.FieldCreatedAt))

	opt := &biz.ListDictTypeOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(dicttype.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(dicttype.DeletedAtIsNil())
	}

	if v := params.Name; len(v) > 0 {
		query = query.Where(dicttype.NameContains(v))
	}
	if v := params.Code; len(v) > 0 {
		query = query.Where(dicttype.CodeContains(v))
	}
	if v := params.Status; len(v) > 0 {
		query = query.Where(dicttype.StatusEQ(dicttype.Status(v)))
	}

	res, err := enthelper.Pagination[*ent.DictType, *ent.DictTypeQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.DictType, 0, res.Total)
	// 转换为biz结构
	for _, v := range res.Data {
		data = append(data, &biz.DictType{
			ID:        v.ID,
			Name:      v.Name,
			Code:      v.Code,
			Status:    string(v.Status),
			Remark:    v.Remark,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			DeletedAt: v.DeletedAt,
		})
	}

	return &biz.ListDictTypeResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

// 创建
func (r *dictTypeRepo) CreateDictType(ctx context.Context, d *biz.DictType) error {
	_, err := r.data.db.DictType.Create().
		SetID(d.ID).
		SetName(d.Name).
		SetCode(d.Code).
		SetStatus(dicttype.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

// 编辑
func (r *dictTypeRepo) UpdateDictType(ctx context.Context, d *biz.DictType) error {
	_, err := r.data.db.DictType.UpdateOneID(d.ID).
		SetName(d.Name).
		SetCode(d.Code).
		SetStatus(dicttype.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

// 删除
func (r *dictTypeRepo) DeleteDictType(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	if err := r.data.db.DictType.
		Update().
		Where(dicttype.IDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx); err != nil {
		return err
	}

	return r.data.db.DictData.
		Update().
		Where(dictdata.DictTypeIDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx)
}

func (r *dictTypeRepo) UpdateDictTypeStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.DictType.UpdateOneID(id).
		SetStatus(dicttype.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
