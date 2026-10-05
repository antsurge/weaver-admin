package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/apiinterface"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menu"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menuapipermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/rolemenu"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

type menuRepo struct {
	data *Data
	log  *log.Helper
}

func NewMenuRepo(data *Data, logger log.Logger) biz.MenuRepo {
	return &menuRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// replaceApiBindings 全量替换菜单绑定的接口（事务内调用）：
// 绑定基于 api_interface.code（service|METHOD|path）关联，不依赖接口 id。
func (r *menuRepo) replaceApiBindings(ctx context.Context, tx *ent.Tx, menuID string, items []*biz.ApiPermission) error {
	// 1. 清空旧绑定
	if _, err := tx.MenuApiPermission.Delete().
		Where(menuapipermission.MenuID(menuID)).
		Exec(ctx); err != nil {
		return err
	}
	// 2. 写入新绑定
	for _, item := range items {
		code := item.CodeKey()
		if code == "" {
			continue
		}
		if _, err := tx.MenuApiPermission.Create().
			SetID(uuid.GenerateXID()).
			SetMenuID(menuID).
			SetAPICode(code).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *menuRepo) CreateMenu(ctx context.Context, p *biz.Menu) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Menu.Create().
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
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx); err != nil {
		return err
	}

	// 绑定接口权限（仅 action 类型携带）
	if err = r.replaceApiBindings(ctx, tx, p.ID, p.APIPermissions); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *menuRepo) UpdateMenu(ctx context.Context, p *biz.Menu) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Menu.UpdateOneID(p.ID).
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
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx); err != nil {
		return err
	}

	// 接口权限全量替换：清空 + 按 code 重新绑定
	if err = r.replaceApiBindings(ctx, tx, p.ID, p.APIPermissions); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteMenus 删除权限及其子孙（事务内：清理角色绑定与接口绑定）
func (r *menuRepo) DeleteMenu(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	var allIDs []string

	// 已访问集合防环：若菜单数据存在循环 parent 引用，直接递归会无限递归导致栈溢出
	visited := make(map[string]bool, len(ids))

	// 事务内递归收集子孙节点
	var collect func(ids []string) error
	collect = func(ids []string) error {
		if len(ids) == 0 {
			return nil
		}

		children, err := tx.Menu.
			Query().
			Where(menu.ParentIDIn(ids...)).
			All(ctx)
		if err != nil {
			return err
		}

		// 收集子节点 ID（去除已访问节点，防止环导致无限递归）
		childIDs := make([]string, 0, len(children))
		for _, c := range children {
			if visited[c.ID] {
				continue
			}
			visited[c.ID] = true
			childIDs = append(childIDs, c.ID)
		}
		allIDs = append(allIDs, childIDs...)

		// 递归查子节点
		return collect(childIDs)
	}

	// 初始化 allIDs 与 visited
	allIDs = append(allIDs, ids...)
	for _, id := range ids {
		visited[id] = true
	}
	if err := collect(ids); err != nil {
		tx.Rollback()
		return err
	}

	// 1. 清理角色-菜单绑定
	_, err = tx.RoleMenu.Delete().
		Where(rolemenu.MenuIDIn(allIDs...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		r.log.Errorf("清理菜单角色绑定失败: %v", err)
		return err
	}

	// 2. 清理菜单-接口绑定（menu_api_permission 表）
	if _, err = tx.MenuApiPermission.Delete().
		Where(menuapipermission.MenuIDIn(allIDs...)).
		Exec(ctx); err != nil {
		tx.Rollback()
		r.log.Errorf("清理菜单接口绑定失败: %v", err)
		return err
	}

	// 3. 删除所有节点
	_, err = tx.Menu.
		Delete().
		Where(menu.IDIn(allIDs...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
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

	if v := params.Type; len(v) > 0 {
		query = query.Where(menu.TypeEQ(menu.Type(v)))
	}

	list, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	bindings, err := r.loadApiBindings(ctx, menuIDsOf(list))
	if err != nil {
		return nil, err
	}

	res := make([]*biz.Menu, 0, len(list))
	for _, v := range list {
		m := r.toBizMenu(v)
		m.APIPermissions = bindings[v.ID]
		res = append(res, m)
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

// GetMenuByID 根据ID查询菜单详情（含接口权限）
func (r *menuRepo) GetMenuByID(ctx context.Context, id string) (*biz.Menu, error) {
	v, err := r.data.db.Menu.Query().
		Where(menu.ID(id)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	bindings, err := r.loadApiBindings(ctx, []string{v.ID})
	if err != nil {
		return nil, err
	}

	m := r.toBizMenu(v)
	m.APIPermissions = bindings[v.ID]
	return m, nil
}

// GetMenusByIDs 根据ID列表查询菜单（用于用户菜单查询）
func (r *menuRepo) GetMenusByIDs(ctx context.Context, ids []string) ([]*biz.Menu, error) {
	if len(ids) == 0 {
		return []*biz.Menu{}, nil
	}

	list, err := r.data.db.Menu.Query().
		Where(menu.IDIn(ids...)).
		Order(ent.Asc(menu.FieldWeight)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	bindings, err := r.loadApiBindings(ctx, menuIDsOf(list))
	if err != nil {
		return nil, err
	}

	res := make([]*biz.Menu, 0, len(list))
	for _, v := range list {
		m := r.toBizMenu(v)
		m.APIPermissions = bindings[v.ID]
		res = append(res, m)
	}

	return res, nil
}

// menuIDsOf 提取菜单 ID 列表
func menuIDsOf(list []*ent.Menu) []string {
	ids := make([]string, 0, len(list))
	for _, v := range list {
		ids = append(ids, v.ID)
	}
	return ids
}

// loadApiBindings 批量加载菜单绑定的接口权限（实时反查 api_interface，不依赖快照）。
// 返回 map[menuID][]*biz.ApiPermission；绑定 code 在当前 api_interface 中不存在时忽略。
func (r *menuRepo) loadApiBindings(ctx context.Context, menuIDs []string) (map[string][]*biz.ApiPermission, error) {
	if len(menuIDs) == 0 {
		return map[string][]*biz.ApiPermission{}, nil
	}

	rows, err := r.data.db.MenuApiPermission.Query().
		Where(menuapipermission.MenuIDIn(menuIDs...)).
		Order(ent.Asc(menuapipermission.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string][]*biz.ApiPermission{}, nil
	}

	// 收集全部绑定 code
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, row.APICode)
	}

	// 实时反查 api_interface（code 唯一，一次查询）
	apis, err := r.data.db.ApiInterface.Query().
		Where(apiinterface.CodeIn(codes...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	apiByCode := make(map[string]*biz.ApiPermission, len(apis))
	for _, a := range apis {
		apiByCode[a.Code] = &biz.ApiPermission{
			Service: a.Service,
			Tag:     a.Tag,
			Method:  a.Method,
			Path:    a.Path,
			Summary: a.Summary,
			Code:    a.Code,
		}
	}

	// 按菜单聚合
	out := make(map[string][]*biz.ApiPermission)
	for _, row := range rows {
		ap, ok := apiByCode[row.APICode]
		if !ok {
			continue
		}
		out[row.MenuID] = append(out[row.MenuID], ap)
	}
	return out, nil
}

// toBizMenu 把 ent.Menu 转成 biz.Menu（接口权限由调用方从 loadApiBindings 注入）
func (r *menuRepo) toBizMenu(v *ent.Menu) *biz.Menu {
	return &biz.Menu{
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
}
