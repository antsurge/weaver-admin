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
	data *Data
	log  *log.Helper
}

func NewMenuRepo(data *Data, logger log.Logger) biz.MenuRepo {
	return &menuRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *menuRepo) CreateMenu(ctx context.Context, p *biz.Menu) error {
	_, err := r.data.db.Menu.Create().
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
		SetBadgeContent(p.BadgeContent).
		SetBadgeStyle(p.BadgeStyle).
		SetWeight(p.Weight).
		SetStatus(menu.Status(p.Status)).
		SetCreatedAt(p.CreatedAt).
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx)
	return err
}

func (r *menuRepo) UpdateMenu(ctx context.Context, p *biz.Menu) error {
	_, err := r.data.db.Menu.UpdateOneID(p.ID).
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
		SetBadgeContent(p.BadgeContent).
		SetBadgeStyle(p.BadgeStyle).
		SetWeight(p.Weight).
		SetStatus(menu.Status(p.Status)).
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx)
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

	list, err := query.All(ctx)

	if err != nil {
		return nil, err
	}

	res := make([]*biz.Menu, 0, len(list))

	for _, v := range list {
		res = append(res, &biz.Menu{
			ID:           v.ID,
			ParentID:     v.ParentID,
			Name:         v.Name,
			Code:         v.Code,
			Title:        v.Title,
			Remark:       v.Remark,
			Path:         v.Path,
			Icon:         v.Icon,
			Type:         string(v.Type),
			LinkUrl:      v.URL,
			Component:    v.Component,
			AuthCode:     v.AuthCode,
			BadgeType:    v.BadgeType,
			BadgeContent: v.BadgeContent,
			BadgeStyle:   v.BadgeStyle,
			Weight:       v.Weight,
			Status:       string(v.Status),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
		})
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
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*biz.Menu, 0, len(list))
	for _, v := range list {
		res = append(res, &biz.Menu{
			ID:           v.ID,
			ParentID:     v.ParentID,
			Name:         v.Name,
			Code:         v.Code,
			Title:        v.Title,
			Remark:       v.Remark,
			Path:         v.Path,
			Icon:         v.Icon,
			Type:         string(v.Type),
			LinkUrl:      v.URL,
			Component:    v.Component,
			AuthCode:     v.AuthCode,
			BadgeType:    v.BadgeType,
			BadgeContent: v.BadgeContent,
			BadgeStyle:   v.BadgeStyle,
			Weight:       v.Weight,
			Status:       string(v.Status),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
		})
	}

	return res, nil
}
