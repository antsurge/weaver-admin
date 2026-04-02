package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/admin"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
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

func (r *adminRepo) ListAdmin(ctx context.Context, params *biz.ListAdminRequest, opts ...*biz.ListAdminOption) (*biz.ListAdminResponse, error) {
	query := r.data.db.Admin.Query().
		Order(ent.Desc(admin.FieldCreatedAt))

	//opt := &biz.ListAdminOption{}
	//if len(opts) > 0 {
	//	opt = opts[0]
	//}

	//if opt.OnlyDeleted {
	//	query = query.Where(admin.DeletedAtNotNil())
	//} else if !opt.IncludeDeleted {
	//	query = query.Where(position.DeletedAtIsNil())
	//}

	res, err := enthelper.Pagination[
		*ent.Admin,
		*ent.AdminQuery,
	](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.Admin, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.Admin{
			ID:        v.ID,
			RealName:  v.RealName,
			Username:  v.Username,
			Email:     v.Email,
			Phone:     v.Phone,
			Avatar:    v.Avatar,
			Password:  v.Password,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		})
	}

	return &biz.ListAdminResponse{
		Items: data,
		Total: res.Total,
	}, nil
}

func (r *adminRepo) CreateAdmin(ctx context.Context, admin *biz.Admin) error {
	_, err := r.data.db.Admin.Create().
		SetID(admin.ID).
		SetRealName(admin.RealName).
		SetUsername(admin.Username).
		SetEmail(admin.Email).
		SetPhone(admin.Phone).
		SetAvatar(admin.Avatar).
		SetPassword(admin.Password).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	return err
}

func (r *adminRepo) UpdateAdmin(ctx context.Context, admin *biz.Admin) error {
	_, err := r.data.db.Admin.
		UpdateOneID(admin.ID).
		SetRealName(admin.RealName).
		SetUsername(admin.Username).
		SetEmail(admin.Email).
		SetPhone(admin.Phone).
		SetAvatar(admin.Avatar).
		SetPassword(admin.Password).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	return err
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
		ID:        entAdmin.ID,
		RealName:  entAdmin.RealName,
		Username:  entAdmin.Username,
		Email:     entAdmin.Email,
		Phone:     entAdmin.Phone,
		Avatar:    entAdmin.Avatar,
		Password:  entAdmin.Password, // 注意：这里包含密码哈希
		CreatedAt: entAdmin.CreatedAt,
		UpdatedAt: entAdmin.UpdatedAt,
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
		ID:        entAdmin.ID,
		RealName:  entAdmin.RealName,
		Username:  entAdmin.Username,
		Email:     entAdmin.Email,
		Phone:     entAdmin.Phone,
		Avatar:    entAdmin.Avatar,
		Password:  entAdmin.Password, // 注意：这里包含密码哈希
		CreatedAt: entAdmin.CreatedAt,
		UpdatedAt: entAdmin.UpdatedAt,
	}, nil
}
