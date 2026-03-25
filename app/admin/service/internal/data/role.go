package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
)

type roleRepo struct {
	data *Data
	log  *log.Helper
}

func NewRoleRepo(data *Data, logger log.Logger) biz.RoleRepo {
	return &roleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *roleRepo) CreateMenu(ctx context.Context, data *biz.Role) error {
	return nil
}
