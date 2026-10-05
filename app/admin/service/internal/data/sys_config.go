package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysconfig"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.ConfigRepo = (*configRepo)(nil)

type configRepo struct {
	data *Data
	log  *log.Helper
}

func NewConfigRepo(data *Data, logger log.Logger) biz.ConfigRepo {
	return &configRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// 列表（支持分页）
func (r *configRepo) ListConfig(ctx context.Context, params *biz.ListConfigRequest, opts ...*biz.ListConfigOption) (*biz.ListConfigResponse, error) {
	query := r.data.db.SysConfig.Query().
		Order(ent.Desc(sysconfig.FieldCreatedAt))

	opt := &biz.ListConfigOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(sysconfig.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(sysconfig.DeletedAtIsNil())
	}

	if v := params.Name; len(v) > 0 {
		query = query.Where(sysconfig.NameContains(v))
	}
	if v := params.Key; len(v) > 0 {
		query = query.Where(sysconfig.KeyContains(v))
	}
	if v := params.Status; len(v) > 0 {
		query = query.Where(sysconfig.StatusEQ(sysconfig.Status(v)))
	}
	if v := params.Group; len(v) > 0 {
		query = query.Where(sysconfig.GroupEQ(v))
	}

	res, err := enthelper.Pagination[*ent.SysConfig, *ent.SysConfigQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.SysConfig, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.SysConfig{
			ID:         v.ID,
			Name:       v.Name,
			Key:        v.Key,
			Value:      v.Value,
			ConfigType: string(v.ConfigType),
			Status:     string(v.Status),
			Remark:     v.Remark,
			Group:      v.Group,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
			DeletedAt:  v.DeletedAt,
		})
	}

	return &biz.ListConfigResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

// 根据键名查询（未删除且启用）
func (r *configRepo) GetConfigByKey(ctx context.Context, key string) (*biz.SysConfig, error) {
	v, err := r.data.db.SysConfig.Query().
		Where(
			sysconfig.KeyEQ(key),
			sysconfig.DeletedAtIsNil(),
			sysconfig.StatusEQ(sysconfig.StatusEnabled),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &biz.SysConfig{
		ID:         v.ID,
		Name:       v.Name,
		Key:        v.Key,
		Value:      v.Value,
		ConfigType: string(v.ConfigType),
		Status:     string(v.Status),
		Remark:     v.Remark,
		Group:      v.Group,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
		DeletedAt:  v.DeletedAt,
	}, nil
}

// 创建
func (r *configRepo) CreateConfig(ctx context.Context, d *biz.SysConfig) error {
	_, err := r.data.db.SysConfig.Create().
		SetID(d.ID).
		SetName(d.Name).
		SetKey(d.Key).
		SetValue(d.Value).
		SetConfigType(sysconfig.ConfigType(d.ConfigType)).
		SetStatus(sysconfig.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetNillableGroup(&d.Group).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

// 编辑
func (r *configRepo) UpdateConfig(ctx context.Context, d *biz.SysConfig) error {
	_, err := r.data.db.SysConfig.UpdateOneID(d.ID).
		SetName(d.Name).
		SetKey(d.Key).
		SetValue(d.Value).
		SetConfigType(sysconfig.ConfigType(d.ConfigType)).
		SetStatus(sysconfig.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetNillableGroup(&d.Group).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

// 删除（软删除）
func (r *configRepo) DeleteConfig(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	return r.data.db.SysConfig.
		Update().
		Where(sysconfig.IDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx)
}

func (r *configRepo) UpdateConfigStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.SysConfig.UpdateOneID(id).
		SetStatus(sysconfig.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
