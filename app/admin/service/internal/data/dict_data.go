package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/dictdata"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
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

func (r *dictDataRepo) List(ctx context.Context, params *biz.ListDictDataRequest, opts ...*biz.ListDictDataOption) (*biz.ListDictDataResponse, error) {
	query := r.data.db.DictData.Query().
		Order(ent.Desc(dictdata.FieldCreatedAt))

	opt := &biz.ListDictDataOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(dictdata.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(dictdata.DeletedAtIsNil())
	}

	if v := params.DictTypeIds; len(v) > 0 {
		query = query.Where(dictdata.DictTypeIDIn(v...))
	}
	if v := params.Label; len(v) > 0 {
		query = query.Where(dictdata.LabelContains(v))
	}
	if v := params.Value; len(v) > 0 {
		query = query.Where(dictdata.ValueContains(v))
	}
	if v := params.Status; len(v) > 0 {
		query = query.Where(dictdata.StatusEQ(dictdata.Status(v)))
	}

	res, err := enthelper.Pagination[*ent.DictData, *ent.DictDataQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.DictData, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.DictData{
			ID:         v.ID,
			DictTypeID: v.DictTypeID,
			Label:      v.Label,
			Value:      v.Value,
			Weight:     v.Weight,
			Status:     string(v.Status),
			Extension:  v.Extension,
			Remark:     v.Remark,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
			DeletedAt:  v.DeletedAt,
		})
	}

	return &biz.ListDictDataResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

func (r *dictDataRepo) Create(ctx context.Context, d *biz.DictData) error {
	_, err := r.data.db.DictData.Create().
		SetID(d.ID).
		SetDictTypeID(d.DictTypeID).
		SetLabel(d.Label).
		SetValue(d.Value).
		SetWeight(d.Weight).
		SetStatus(dictdata.Status(d.Status)).
		SetExtension(d.Extension).
		SetNillableRemark(&d.Remark).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *dictDataRepo) Update(ctx context.Context, d *biz.DictData) error {
	_, err := r.data.db.DictData.UpdateOneID(d.ID).
		SetDictTypeID(d.DictTypeID).
		SetLabel(d.Label).
		SetValue(d.Value).
		SetWeight(d.Weight).
		SetStatus(dictdata.Status(d.Status)).
		SetExtension(d.Extension).
		SetNillableRemark(&d.Remark).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *dictDataRepo) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	return r.data.db.DictData.
		Update().
		Where(dictdata.IDIn(ids...)).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}

func (r *dictDataRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.DictData.UpdateOneID(id).
		SetStatus(dictdata.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func ExtensionFromJSONString(value string) (map[string]any, error) {
	if value == "" {
		return map[string]any{}, nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}
	return result, nil
}
