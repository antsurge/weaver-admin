package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/admin"
)

type adminRepo struct {
	data *Data
	log  *log.Helper
}

func NewAdminRepo(data *Data, logger log.Logger) biz.AdminRepo {
	return &adminRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *adminRepo) CreateAdmin(ctx context.Context, a *biz.Admin) (*biz.Admin, error) {
	// 你的实现逻辑...
	return nil, nil
}

// FindByUsername 根据用户名查找管理员
func (r *adminRepo) FindByUsername(ctx context.Context, username string) (*biz.Admin, error) {
	entAdmin, err := r.data.db.Admin.Query().
		Where(admin.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		r.log.Errorf("failed to query admin by username: %v", err)
		return nil, err
	}

	return &biz.Admin{
		ID:         entAdmin.ID,
		RealName:   entAdmin.RealName,
		Username:   entAdmin.Username,
		Email:      entAdmin.Email,
		Phone:      entAdmin.Phone,
		Avatar:     entAdmin.Avatar,
		Password:   entAdmin.Password, // 注意：这里包含密码哈希
		CreateTime: entAdmin.CreateTime,
		UpdateTime: entAdmin.UpdateTime,
	}, nil
}

// FindByUsername 根据用户名查找管理员
func (r *adminRepo) FindByID(ctx context.Context, id string) (*biz.Admin, error) {
	entAdmin, err := r.data.db.Admin.Query().
		Where(admin.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		r.log.Errorf("failed to query admin by id: %v", err)
		return nil, err
	}

	return &biz.Admin{
		ID:         entAdmin.ID,
		RealName:   entAdmin.RealName,
		Username:   entAdmin.Username,
		Email:      entAdmin.Email,
		Phone:      entAdmin.Phone,
		Avatar:     entAdmin.Avatar,
		Password:   entAdmin.Password, // 注意：这里包含密码哈希
		CreateTime: entAdmin.CreateTime,
		UpdateTime: entAdmin.UpdateTime,
	}, nil
}
