package data

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/adminrole"
	"github.com/go-kratos/kratos/v2/log"
)

type adminRoleRepo struct {
	data *Data
	log  *log.Helper
}

func NewAdminRoleRepo(data *Data, logger log.Logger) biz.AdminRoleRepo {
	return &adminRoleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *adminRoleRepo) GetRoleIdsByAdminId(ctx context.Context, id string) ([]string, error) {
	roleIds, err := r.data.db.AdminRole.
		Query().
		Where(
			adminrole.AdminID(id),
		).
		Select(adminrole.FieldRoleID).
		Strings(ctx)

	if err != nil {
		return nil, err
	}

	return roleIds, nil
}
