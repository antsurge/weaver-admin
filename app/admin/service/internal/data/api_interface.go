package data

import (
	"context"
	"sort"

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

// ListApiInterfaceOptions 查询去重后的服务名/标签（下拉数据源）
func (r *apiInterfaceRepo) ListApiInterfaceOptions(ctx context.Context) (*biz.ApiInterfaceOptions, error) {
	services, err := r.data.db.ApiInterface.Query().
		Where(apiinterface.ServiceNEQ("")).
		GroupBy(apiinterface.FieldService).
		Strings(ctx)
	if err != nil {
		return nil, err
	}

	tags, err := r.data.db.ApiInterface.Query().
		Where(apiinterface.TagNEQ("")).
		GroupBy(apiinterface.FieldTag).
		Strings(ctx)
	if err != nil {
		return nil, err
	}

	sort.Strings(services)
	sort.Strings(tags)
	return &biz.ApiInterfaceOptions{Services: services, Tags: tags}, nil
}

// GetApiInterfaceByCode 按唯一键查询（用于新增/编辑时冲突检查；不存在返回 nil）
func (r *apiInterfaceRepo) GetApiInterfaceByCode(ctx context.Context, code string) (*biz.ApiInterface, error) {
	v, err := r.data.db.ApiInterface.Query().
		Where(apiinterface.CodeEQ(code)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizApiInterface(v), nil
}

// CreateApiInterface 手动新增
func (r *apiInterfaceRepo) CreateApiInterface(ctx context.Context, item *biz.ApiInterface) (*biz.ApiInterface, error) {
	v, err := r.data.db.ApiInterface.Create().
		SetID(item.ID).
		SetService(item.Service).
		SetTag(item.Tag).
		SetMethod(item.Method).
		SetPath(item.Path).
		SetSummary(item.Summary).
		SetCode(item.Code).
		SetCreatedAt(item.CreatedAt).
		SetUpdatedAt(item.UpdatedAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toBizApiInterface(v), nil
}

// UpdateApiInterface 手动编辑
func (r *apiInterfaceRepo) UpdateApiInterface(ctx context.Context, item *biz.ApiInterface) (*biz.ApiInterface, error) {
	v, err := r.data.db.ApiInterface.UpdateOneID(item.ID).
		SetService(item.Service).
		SetTag(item.Tag).
		SetMethod(item.Method).
		SetPath(item.Path).
		SetSummary(item.Summary).
		SetCode(item.Code).
		SetUpdatedAt(item.UpdatedAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toBizApiInterface(v), nil
}

// UpsertApiInterfaces 按 code 对比导入（事务保证原子性）：
//   - code 已存在的记录做更新（保留原 id，绑定关系不受影响，因为绑定按 code 关联）
//   - code 不存在的记录做创建
//
// 不会删除任何已有数据，手动新增的接口不受影响。
func (r *apiInterfaceRepo) UpsertApiInterfaces(ctx context.Context, items []*biz.ApiInterface) (*biz.ImportResult, error) {
	result := &biz.ImportResult{Total: len(items)}

	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 查出当前全部 code -> id 映射
	codes, err := tx.ApiInterface.Query().
		Select(apiinterface.FieldCode, apiinterface.FieldID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	codeToID := make(map[string]string, len(codes))
	for _, c := range codes {
		codeToID[c.Code] = c.ID
	}

	for _, item := range items {
		if existingID, ok := codeToID[item.Code]; ok {
			// 已存在 -> 更新
			if _, err = tx.ApiInterface.UpdateOneID(existingID).
				SetService(item.Service).
				SetTag(item.Tag).
				SetMethod(item.Method).
				SetPath(item.Path).
				SetSummary(item.Summary).
				SetCode(item.Code).
				SetUpdatedAt(item.UpdatedAt).
				Save(ctx); err != nil {
				return nil, err
			}
			result.Updated++
		} else {
			// 不存在 -> 创建
			if _, err = tx.ApiInterface.Create().
				SetID(item.ID).
				SetService(item.Service).
				SetTag(item.Tag).
				SetMethod(item.Method).
				SetPath(item.Path).
				SetSummary(item.Summary).
				SetCode(item.Code).
				SetCreatedAt(item.CreatedAt).
				SetUpdatedAt(item.UpdatedAt).
				Save(ctx); err != nil {
				return nil, err
			}
			result.Imported++
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
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
