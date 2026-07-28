package data

import (
	"context"
	"errors"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/apipermission"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

type apiPermissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewApiPermissionRepo(data *Data, logger log.Logger) biz.ApiPermissionRepo {
	return &apiPermissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// UpsertByCodes 按业务唯一键 code（service|method|path）创建或查找 ApiPermission 记录，
// 返回带 ID 的 biz 实体供 m2m 关系绑定使用。
func (r *apiPermissionRepo) UpsertByCodes(ctx context.Context, items []*biz.ApiPermission) ([]*biz.ApiPermission, error) {
	if len(items) == 0 {
		return nil, nil
	}

	results := make([]*biz.ApiPermission, 0, len(items))
	for _, item := range items {
		code := item.Code()
		// 查找现有
		exist, err := r.data.db.ApiPermission.Query().
			Where(apipermission.CodeEQ(code)).
			First(ctx)
		if err == nil {
			results = append(results, &biz.ApiPermission{
				ID:      exist.ID,
				Service: exist.Service,
				Method:  exist.Method,
				Path:    exist.Path,
				Summary: exist.Summary,
			})
			continue
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
		// 不存在则创建
		summary := item.Summary
		created, err := r.data.db.ApiPermission.Create().
			SetID(uuid.GenerateXID()).
			SetService(item.Service).
			SetMethod(item.Method).
			SetPath(item.Path).
			SetSummary(summary).
			SetCode(code).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, &biz.ApiPermission{
			ID:      created.ID,
			Service: created.Service,
			Method:  created.Method,
			Path:    created.Path,
			Summary: created.Summary,
		})
	}
	return results, nil
}

// ListByCodes 按 code 列表查询（用于已绑定接口权限回显）
func (r *apiPermissionRepo) ListByCodes(ctx context.Context, codes []string) ([]*biz.ApiPermission, error) {
	if len(codes) == 0 {
		return []*biz.ApiPermission{}, nil
	}
	list, err := r.data.db.ApiPermission.Query().
		Where(apipermission.CodeIn(codes...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*biz.ApiPermission, 0, len(list))
	for _, v := range list {
		out = append(out, &biz.ApiPermission{
			ID:      v.ID,
			Service: v.Service,
			Method:  v.Method,
			Path:    v.Path,
			Summary: v.Summary,
		})
	}
	return out, nil
}

// ListAll 列出所有已注册的接口（用于扫描器对账/管理）
func (r *apiPermissionRepo) ListAll(ctx context.Context) ([]*biz.ApiPermission, error) {
	list, err := r.data.db.ApiPermission.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*biz.ApiPermission, 0, len(list))
	for _, v := range list {
		out = append(out, &biz.ApiPermission{
			ID:      v.ID,
			Service: v.Service,
			Method:  v.Method,
			Path:    v.Path,
			Summary: v.Summary,
		})
	}
	return out, nil
}

var _ = errors.New