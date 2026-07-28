package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menu"
	"github.com/go-kratos/kratos/v2/log"
)

type menuRepo struct {
	data    *Data
	log     *log.Helper
	apiPerm biz.ApiPermissionRepo
}

func NewMenuRepo(data *Data, logger log.Logger, apiPerm biz.ApiPermissionRepo) biz.MenuRepo {
	return &menuRepo{
		data:    data,
		log:     log.NewHelper(logger),
		apiPerm: apiPerm,
	}
}

// resolveAPIPermissions 将 biz.ApiPermission 列表 upsert 到数据库并返回带 ID 的 biz 实体
func (r *menuRepo) resolveAPIPermissions(ctx context.Context, items []*biz.ApiPermission) ([]*biz.ApiPermission, error) {
	if len(items) == 0 {
		return nil, nil
	}
	return r.apiPerm.UpsertByCodes(ctx, items)
}

func (r *menuRepo) CreateMenu(ctx context.Context, p *biz.Menu) error {
	builder := r.data.db.Menu.Create().
		SetID(p.ID).
		SetParentID(p.ParentID).
		SetName(p.Name).
		SetCode(p.Code).
		SetTitle(p.Title).
		SetNillableRemark(&p.Remark).
		SetPath(p.Path).
		SetIcon(p.Icon).
		SetType(menu.Type(p.Type)).
		SetURL(p.LinkUrl).
		SetComponent(p.Component).
		SetAuthCode(p.AuthCode).
		SetBadgeType(p.BadgeType).
		SetBadge(p.Badge).
		SetBadgeVariants(p.BadgeVariants).
		SetWeight(p.Weight).
		SetStatus(menu.Status(p.Status)).
		SetCreatedAt(p.CreatedAt).
		SetUpdatedAt(p.UpdatedAt)

	// 绑定接口权限（仅 action 类型携带）
	if len(p.APIPermissions) > 0 {
		ents, err := r.resolveAPIPermissions(ctx, p.APIPermissions)
		if err != nil {
			return err
		}
		ids := make([]string, len(ents))
		for i, e := range ents {
			ids[i] = e.ID
		}
		builder = builder.AddAPIPermissionIDs(ids...)
	}

	_, err := builder.Save(ctx)
	return err
}

func (r *menuRepo) UpdateMenu(ctx context.Context, p *biz.Menu) error {
	builder := r.data.db.Menu.UpdateOneID(p.ID).
		SetParentID(p.ParentID).
		SetName(p.Name).
		SetCode(p.Code).
		SetTitle(p.Title).
		SetNillableRemark(&p.Remark).
		SetPath(p.Path).
		SetIcon(p.Icon).
		SetType(menu.Type(p.Type)).
		SetURL(p.LinkUrl).
		SetComponent(p.Component).
		SetAuthCode(p.AuthCode).
		SetBadgeType(p.BadgeType).
		SetBadge(p.Badge).
		SetBadgeVariants(p.BadgeVariants).
		SetWeight(p.Weight).
		SetStatus(menu.Status(p.Status)).
		SetUpdatedAt(p.UpdatedAt)

	// 接口权限全量替换：清空 + 重新绑定
	ents, err := r.resolveAPIPermissions(ctx, p.APIPermissions)
	if err != nil {
		return err
	}
	ids := make([]string, len(ents))
	for i, e := range ents {
		ids[i] = e.ID
	}
	builder = builder.ClearAPIPermissions().AddAPIPermissionIDs(ids...)

	_, err = builder.Save(ctx)
	return err
}

// DeleteMenus 删除权限及其子孙
func (r *menuRepo) DeleteMenu(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var allIDs []string

	// 递归收集子孙节点
	var collect func(ids []string) error
	collect = func(ids []string) error {
		if len(ids) == 0 {
			return nil
		}

		children, err := r.data.db.Menu.
			Query().
			Where(menu.ParentIDIn(ids...)).
			All(ctx)
		if err != nil {
			return err
		}

		// 收集子节点 ID
		childIDs := make([]string, len(children))
		for i, c := range children {
			childIDs[i] = c.ID
		}
		allIDs = append(allIDs, childIDs...)

		// 递归查子节点
		return collect(childIDs)
	}

	// 初始化 allIDs
	allIDs = append(allIDs, ids...)
	if err := collect(ids); err != nil {
		return err
	}

	// 删除所有节点
	_, err := r.data.db.Menu.
		Delete().
		Where(menu.IDIn(allIDs...)).
		Exec(ctx)
	return err
}

func (r *menuRepo) ListMenu(
	ctx context.Context,
	params *biz.ListMenuRequest,
) ([]*biz.Menu, error) {
	query := r.data.db.Menu.Query().
		Order(ent.Desc(menu.FieldWeight))

	if v := params.Name; len(v) > 0 {
		query = query.Where(menu.NameContains(v))
	}

	if v := params.Code; len(v) > 0 {
		query = query.Where(menu.CodeContains(v))
	}

	if v := params.Status; len(v) > 0 {
		query = query.Where(menu.StatusEQ(menu.Status(v)))
	}

	list, err := query.
		WithAPIPermissions().
		All(ctx)

	if err != nil {
		return nil, err
	}

	res := make([]*biz.Menu, 0, len(list))

	for _, v := range list {
		res = append(res, r.toBizMenu(v))
	}

	return res, nil
}

func (r *menuRepo) UpdateMenuStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Menu.UpdateOneID(id).
		SetStatus(menu.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// GetMenusByIDs 根据ID列表查询菜单（用于用户菜单查询）
func (r *menuRepo) GetMenusByIDs(ctx context.Context, ids []string) ([]*biz.Menu, error) {
	if len(ids) == 0 {
		return []*biz.Menu{}, nil
	}

	list, err := r.data.db.Menu.Query().
		Where(menu.IDIn(ids...)).
		Order(ent.Asc(menu.FieldWeight)).
		WithAPIPermissions().
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*biz.Menu, 0, len(list))
	for _, v := range list {
		res = append(res, r.toBizMenu(v))
	}

	return res, nil
}

// toBizMenu 把 ent.Menu 转成 biz.Menu（含 apiPermissions 转换）
func (r *menuRepo) toBizMenu(v *ent.Menu) *biz.Menu {
	m := &biz.Menu{
		ID:            v.ID,
		ParentID:      v.ParentID,
		Name:          v.Name,
		Code:          v.Code,
		Title:         v.Title,
		Remark:        v.Remark,
		Path:          v.Path,
		Icon:          v.Icon,
		Type:          string(v.Type),
		LinkUrl:       v.URL,
		Component:     v.Component,
		AuthCode:      v.AuthCode,
		BadgeType:     v.BadgeType,
		Badge:         v.Badge,
		BadgeVariants: v.BadgeVariants,
		Weight:        v.Weight,
		Status:        string(v.Status),
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
	for _, ap := range v.Edges.APIPermissions {
		m.APIPermissions = append(m.APIPermissions, &biz.ApiPermission{
			ID:      ap.ID,
			Service: ap.Service,
			Method:  ap.Method,
			Path:    ap.Path,
			Summary: ap.Summary,
		})
	}
	return m
}
