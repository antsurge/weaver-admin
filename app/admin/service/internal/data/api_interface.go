package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/apiinterface"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.ApiInterfaceRepo = (*apiInterfaceRepo)(nil)

type apiInterfaceRepo struct {
	data *Data
	log  *log.Helper
}

func NewApiInterfaceRepo(data *Data, logger log.Logger) biz.ApiInterfaceRepo {
	return &apiInterfaceRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// ListApiInterface 分页查询
func (r *apiInterfaceRepo) ListApiInterface(ctx context.Context, params *biz.ListApiInterfaceRequest) (*biz.ListApiInterfaceResponse, error) {
	query := r.data.db.ApiInterface.Query().
		Order(ent.Desc(apiinterface.FieldCreatedAt))

	if v := params.Service; len(v) > 0 {
		query = query.Where(apiinterface.ServiceContains(v))
	}
	if v := params.Tag; len(v) > 0 {
		query = query.Where(apiinterface.TagContains(v))
	}
	if v := params.Method; len(v) > 0 {
		query = query.Where(apiinterface.MethodContains(v))
	}
	if v := params.Path; len(v) > 0 {
		query = query.Where(apiinterface.PathContains(v))
	}
	if v := params.Summary; len(v) > 0 {
		query = query.Where(apiinterface.SummaryContains(v))
	}

	res, err := enthelper.Pagination[*ent.ApiInterface, *ent.ApiInterfaceQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.ApiInterface, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, toBizApiInterface(v))
	}

	return &biz.ListApiInterfaceResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

// UpsertApiInterface 按 code 插入或更新
func (r *apiInterfaceRepo) UpsertApiInterface(ctx context.Context, d *biz.ApiInterface) error {
	// 先查是否存在
	exists, err := r.data.db.ApiInterface.Query().
		Where(apiinterface.CodeEQ(d.Code)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return err
		}
		// 不存在则创建
		_, err = r.data.db.ApiInterface.Create().
			SetID(d.ID).
			SetService(d.Service).
			SetTag(d.Tag).
			SetMethod(d.Method).
			SetPath(d.Path).
			SetSummary(d.Summary).
			SetCode(d.Code).
			SetCreatedAt(d.CreatedAt).
			SetUpdatedAt(d.UpdatedAt).
			Save(ctx)
		return err
	}

	// 存在则更新
	_, err = r.data.db.ApiInterface.UpdateOneID(exists.ID).
		SetService(d.Service).
		SetTag(d.Tag).
		SetMethod(d.Method).
		SetPath(d.Path).
		SetSummary(d.Summary).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// DeleteApiInterface 批量删除
func (r *apiInterfaceRepo) DeleteApiInterface(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.data.db.ApiInterface.Delete().
		Where(apiinterface.IDIn(ids...)).
		Exec(ctx)
	return err
}

// toBizApiInterface ent -> biz 转换
func toBizApiInterface(v *ent.ApiInterface) *biz.ApiInterface {
	return &biz.ApiInterface{
		ID:        v.ID,
		Service:   v.Service,
		Tag:       v.Tag,
		Method:    v.Method,
		Path:      v.Path,
		Summary:   v.Summary,
		Code:      v.Code,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}