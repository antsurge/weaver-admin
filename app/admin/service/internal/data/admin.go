package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/admin"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/adminrole"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
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

	opt := &biz.ListAdminOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(admin.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(admin.DeletedAtIsNil())
	}

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
		adminItem := &biz.Admin{
			ID:        v.ID,
			RealName:  v.RealName,
			Username:  v.Username,
			Email:     v.Email,
			Phone:     v.Phone,
			Avatar:    v.Avatar,
			Password:  v.Password,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}

		// 填充角色ID列表
		roleIDs, err := r.GetRoleIDsByAdmin(ctx, v.ID)
		if err != nil {
			r.log.Warnf("获取用户 %s 的角色失败: %v", v.ID, err)
			roleIDs = []string{}
		}
		adminItem.RoleIDs = roleIDs

		data = append(data, adminItem)
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
		Where(
			admin.UsernameEQ(username),
			admin.DeletedAtIsNil(),
		).
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
		Password:  entAdmin.Password,
		CreatedAt: entAdmin.CreatedAt,
		UpdatedAt: entAdmin.UpdatedAt,
	}, nil
}

// FindByID 根据ID查找管理员
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
		Password:  entAdmin.Password,
		CreatedAt: entAdmin.CreatedAt,
		UpdatedAt: entAdmin.UpdatedAt,
	}, nil
}

// DeleteAdmin 删除用户（软删除，设置 deleted_at 时间）
func (r *adminRepo) DeleteAdmin(ctx context.Context, ids []string) error {
	err := r.data.db.Admin.
		Update().
		Where(admin.IDIn(ids...)).
		SetDeletedAt(time.Now()).
		Exec(ctx)

	return err
}

// ====== 角色关联方法实现 ======

// BindRolesForAdmin 为用户绑定角色（全量替换：先删除所有现有绑定，再批量插入新绑定）
func (r *adminRepo) BindRolesForAdmin(ctx context.Context, adminID string, roleIDs []string) error {
	// 开启事务
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	// 1. 删除该用户的所有现有绑定
	_, err = tx.AdminRole.Delete().
		Where(adminrole.AdminIDEQ(adminID)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		r.log.Errorf("删除用户 %s 的旧角色绑定失败: %v", adminID, err)
		return err
	}

	// 2. 批量插入新的绑定（如果有）
	if len(roleIDs) > 0 {
		builders := make([]*ent.AdminRoleCreate, len(roleIDs))
		for i, roleID := range roleIDs {
			builders[i] = tx.AdminRole.Create().
				SetID(uuid.GenerateXID()).
				SetAdminID(adminID).
				SetRoleID(roleID)
		}

		err = tx.AdminRole.CreateBulk(builders...).Exec(ctx)
		if err != nil {
			tx.Rollback()
			r.log.Errorf("为用户 %s 绑定角色失败: %v", adminID, err)
			return err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		r.log.Errorf("提交用户角色绑定事务失败: %v", err)
		return err
	}

	r.log.Infof("用户 %s 绑定了 %d 个角色", adminID, len(roleIDs))
	return nil
}

// GetRoleIDsByAdmin 获取用户关联的角色ID列表
func (r *adminRepo) GetRoleIDsByAdmin(ctx context.Context, adminID string) ([]string, error) {
	// 查询关联表
	roles, err := r.data.db.AdminRole.Query().
		Where(adminrole.AdminIDEQ(adminID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 提取角色ID
	ids := make([]string, len(roles))
	for i, r := range roles {
		ids[i] = r.RoleID
	}

	return ids, nil
}
