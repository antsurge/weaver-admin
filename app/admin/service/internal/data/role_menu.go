package data

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menu"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/rolemenu"
	"github.com/go-kratos/kratos/v2/log"
)

type roleMenuRepo struct {
	data *Data
	log  *log.Helper
}

func NewRoleMenuRepo(data *Data, logger log.Logger) biz.RoleMenuRepo {
	return &roleMenuRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *roleMenuRepo) GetMenuIdsByRoleIds(ctx context.Context, roleIds []string) ([]string, error) {
	if len(roleIds) == 0 {
		return []string{}, nil
	}

	menuIds, err := r.data.db.RoleMenu.
		Query().
		Where(
			rolemenu.RoleIDIn(roleIds...),
		).
		QueryMenu().
		Where(
			menu.StatusEQ("enabled"),
		).
		IDs(ctx)

	if err != nil {
		return nil, err
	}

	return menuIds, nil
}
