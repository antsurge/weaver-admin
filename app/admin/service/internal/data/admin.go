package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	entadmin "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/admin"
	entadmindataprm "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/admindatapermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/adminrole"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/datapermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/department"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/predicate"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/role"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
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
		Order(ent.Desc(entadmin.FieldCreatedAt))

	opt := &biz.ListAdminOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(entadmin.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(entadmin.DeletedAtIsNil())
	}

	// 过滤条件（模糊匹配用户名/真实姓名/手机号/邮箱，精确匹配状态）
	if params.Username != "" {
		query = query.Where(entadmin.UsernameContainsFold(params.Username))
	}
	if params.RealName != "" {
		query = query.Where(entadmin.RealNameContainsFold(params.RealName))
	}
	if params.Phone != "" {
		query = query.Where(entadmin.PhoneContainsFold(params.Phone))
	}
	if params.Email != "" {
		query = query.Where(entadmin.EmailContainsFold(params.Email))
	}
	if params.Status != "" {
		query = query.Where(entadmin.StatusEQ(entadmin.Status(params.Status)))
	}
	// 归属部门（精确匹配）
	if params.DepartmentID != "" {
		query = query.Where(entadmin.DepartmentIDEQ(params.DepartmentID))
	}

	res, err := enthelper.Pagination[
		*ent.Admin,
		*ent.AdminQuery,
	](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构，并批量填充角色ID列表（一次查询，避免 N+1）
	adminIDs := make([]string, 0, len(res.Data))
	for _, v := range res.Data {
		adminIDs = append(adminIDs, v.ID)
	}
	roleIDsByAdmin, err := r.getRoleIDsByAdmins(ctx, adminIDs)
	if err != nil {
		r.log.Warnf("批量获取用户角色失败: %v", err)
		roleIDsByAdmin = map[string][]string{}
	}

	// 批量获取用户绑定的数据权限规则（一次查询，避免 N+1）
	dpIDsByAdmin, err := r.getDataPermissionIDsByAdmins(ctx, adminIDs)
	if err != nil {
		r.log.Warnf("批量获取用户数据权限规则失败: %v", err)
		dpIDsByAdmin = map[string][]string{}
	}

	// 收集全部数据权限规则ID，一次性查出规则名，避免 N+1
	allDPIDs := make([]string, 0)
	for _, ids := range dpIDsByAdmin {
		allDPIDs = append(allDPIDs, ids...)
	}
	dpNameByID := make(map[string]string)
	if len(allDPIDs) > 0 {
		dpRuleNames, err := r.getDataPermissionNamesByIDs(ctx, allDPIDs)
		if err != nil {
			r.log.Warnf("批量获取数据权限规则名称失败: %v", err)
		} else {
			dpNameByID = dpRuleNames
		}
	}

	// 收集全部角色ID，一次性查出角色名，避免 N+1
	allRoleIDs := make([]string, 0)
	for _, ids := range roleIDsByAdmin {
		allRoleIDs = append(allRoleIDs, ids...)
	}
	roleNameByID, err := r.getRoleNamesByIDs(ctx, allRoleIDs)
	if err != nil {
		r.log.Warnf("批量获取角色名称失败: %v", err)
		roleNameByID = map[string]string{}
	}

	// 收集全部归属部门ID，一次性查出部门名，避免 N+1
	allDeptIDs := make([]string, 0)
	deptIDSet := make(map[string]struct{}, len(res.Data))
	for _, v := range res.Data {
		if v.DepartmentID == "" {
			continue
		}
		if _, ok := deptIDSet[v.DepartmentID]; ok {
			continue
		}
		deptIDSet[v.DepartmentID] = struct{}{}
		allDeptIDs = append(allDeptIDs, v.DepartmentID)
	}
	deptNameByID, err := r.getDepartmentNamesByIDs(ctx, allDeptIDs)
	if err != nil {
		r.log.Warnf("批量获取部门名称失败: %v", err)
		deptNameByID = map[string]string{}
	}

	data := make([]*biz.Admin, 0, res.Total)
	for _, v := range res.Data {
		adminItem := &biz.Admin{
			ID:           v.ID,
			TenantID:     v.TenantID,
			RealName:     v.RealName,
			Username:     v.Username,
			Email:        v.Email,
			Phone:        v.Phone,
			Avatar:       v.Avatar,
			Password:     v.Password,
			Status:       string(v.Status),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
			DepartmentID: v.DepartmentID,
			RoleIDs:      roleIDsByAdmin[v.ID],
		}
		adminItem.DepartmentName = deptNameByID[v.DepartmentID]
		if adminItem.RoleIDs == nil {
			adminItem.RoleIDs = []string{}
		}
		adminItem.RoleNames = make([]string, 0, len(adminItem.RoleIDs))
		for _, roleID := range adminItem.RoleIDs {
			if name, ok := roleNameByID[roleID]; ok {
				adminItem.RoleNames = append(adminItem.RoleNames, name)
			}
		}

		// 数据权限规则：ID列表与名称列表（空 = 跟随角色）
		adminItem.DataPermissionIDs = dpIDsByAdmin[v.ID]
		if adminItem.DataPermissionIDs == nil {
			adminItem.DataPermissionIDs = []string{}
		}
		adminItem.DataPermissionNames = make([]string, 0, len(adminItem.DataPermissionIDs))
		for _, dpID := range adminItem.DataPermissionIDs {
			if name, ok := dpNameByID[dpID]; ok {
				adminItem.DataPermissionNames = append(adminItem.DataPermissionNames, name)
			}
		}

		data = append(data, adminItem)
	}

	return &biz.ListAdminResponse{
		Items: data,
		Total: res.Total,
	}, nil
}

func (r *adminRepo) CreateAdmin(ctx context.Context, admin *biz.Admin) error {
	createOne := r.data.db.Admin.Create().
		SetID(admin.ID).
		SetRealName(admin.RealName).
		SetUsername(admin.Username).
		SetEmail(admin.Email).
		SetPhone(admin.Phone).
		SetAvatar(admin.Avatar).
		SetPassword(admin.Password).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now())

	// 归属部门（可选，空字符串保持默认值）
	if admin.DepartmentID != "" {
		createOne = createOne.SetDepartmentID(admin.DepartmentID)
	}

	_, err := createOne.SetStatus(entadmin.Status(admin.Status)).Save(ctx)

	return err
}

func (r *adminRepo) UpdateAdmin(ctx context.Context, admin *biz.Admin) error {
	updateOne := r.data.db.Admin.
		UpdateOneID(admin.ID).
		SetRealName(admin.RealName).
		SetUsername(admin.Username).
		SetEmail(admin.Email).
		SetPhone(admin.Phone).
		SetAvatar(admin.Avatar).
		SetDepartmentID(admin.DepartmentID).
		SetUpdatedAt(time.Now())

	// 密码非空才更新（空字符串表示不修改密码）
	if admin.Password != "" {
		updateOne = updateOne.SetPassword(admin.Password)
	}

	// 状态非空才更新（UpdateAdminRequest.status 可选）
	if admin.Status != "" {
		updateOne = updateOne.SetStatus(entadmin.Status(admin.Status))
	}

	_, err := updateOne.Save(ctx)
	return err
}

// FindByUsername 根据用户名查找管理员
func (r *adminRepo) FindByUsername(ctx context.Context, username string) (*biz.Admin, error) {
	entAdmin, err := r.data.db.Admin.Query().
		Where(
			entadmin.UsernameEQ(username),
			entadmin.DeletedAtIsNil(),
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
		ID:           entAdmin.ID,
		TenantID:     entAdmin.TenantID,
		RealName:     entAdmin.RealName,
		Username:     entAdmin.Username,
		Email:        entAdmin.Email,
		Phone:        entAdmin.Phone,
		Avatar:       entAdmin.Avatar,
		Password:     entAdmin.Password,
		Status:       string(entAdmin.Status),
		CreatedAt:    entAdmin.CreatedAt,
		UpdatedAt:    entAdmin.UpdatedAt,
		DepartmentID: entAdmin.DepartmentID,
	}, nil
}

// FindByID 根据ID查找管理员（已软删除的记录不可见）
func (r *adminRepo) FindByID(ctx context.Context, id string) (*biz.Admin, error) {
	entAdmin, err := r.data.db.Admin.Query().
		Where(
			entadmin.IDEQ(id),
			entadmin.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		r.log.Errorf("failed to query admin by id: %v", err)
		return nil, err
	}

	out := &biz.Admin{
		ID:           entAdmin.ID,
		TenantID:     entAdmin.TenantID,
		RealName:     entAdmin.RealName,
		Username:     entAdmin.Username,
		Email:        entAdmin.Email,
		Phone:        entAdmin.Phone,
		Avatar:       entAdmin.Avatar,
		Password:     entAdmin.Password,
		Status:       string(entAdmin.Status),
		CreatedAt:    entAdmin.CreatedAt,
		UpdatedAt:    entAdmin.UpdatedAt,
		DepartmentID: entAdmin.DepartmentID,
	}
	r.fillDepartmentName(ctx, out)
	return out, nil
}

// FindByIDs 批量按ID查询管理员（已软删除的记录不可见）
func (r *adminRepo) FindByIDs(ctx context.Context, ids []string) ([]*biz.Admin, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	list, err := r.data.db.Admin.Query().
		Where(
			entadmin.IDIn(ids...),
			entadmin.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		r.log.Errorf("failed to query admins by ids: %v", err)
		return nil, err
	}

	result := make([]*biz.Admin, 0, len(list))
	for _, v := range list {
		result = append(result, &biz.Admin{
			ID:           v.ID,
			TenantID:     v.TenantID,
			RealName:     v.RealName,
			Username:     v.Username,
			Email:        v.Email,
			Phone:        v.Phone,
			Avatar:       v.Avatar,
			Status:       string(v.Status),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
			DepartmentID: v.DepartmentID,
		})
	}
	return result, nil
}

// DeleteAdmin 批量软删除用户（事务内：清理角色绑定；保护绑定超管角色的用户）
func (r *adminRepo) DeleteAdmin(ctx context.Context, ids []string, superAdminCode string) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	// 1. 绑定超级管理员角色的用户禁止删除
	if superAdminCode != "" {
		bindings, err := tx.AdminRole.Query().
			Where(
				adminrole.AdminIDIn(ids...),
				adminrole.HasRoleWith(role.CodeEQ(superAdminCode)),
			).
			All(ctx)
		if err != nil {
			tx.Rollback()
			return err
		}
		if len(bindings) > 0 {
			tx.Rollback()
			return errors.BadRequest("CANNOT_DELETE_SUPERADMIN", "超级管理员账号不允许删除")
		}
	}

	// 2. 清理角色绑定
	_, err = tx.AdminRole.Delete().
		Where(adminrole.AdminIDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		r.log.Errorf("清理用户角色绑定失败: %v", err)
		return err
	}

	// 2.1 清理数据权限规则绑定
	_, err = tx.AdminDataPermission.Delete().
		Where(entadmindataprm.AdminIDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		r.log.Errorf("清理用户数据权限规则绑定失败: %v", err)
		return err
	}

	// 3. 软删除用户
	err = tx.Admin.Update().
		Where(entadmin.IDIn(ids...)).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdatePassword 更新指定用户密码（biz 层已哈希）
func (r *adminRepo) UpdatePassword(ctx context.Context, adminID, hashedPassword string) error {
	_, err := r.data.db.Admin.
		UpdateOneID(adminID).
		SetPassword(hashedPassword).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// UpdateStatus 更新指定用户状态
func (r *adminRepo) UpdateStatus(ctx context.Context, adminID, status string) error {
	_, err := r.data.db.Admin.
		UpdateOneID(adminID).
		SetStatus(entadmin.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// BatchUpdateStatus 批量更新用户状态（仅作用于未软删除记录）
func (r *adminRepo) BatchUpdateStatus(ctx context.Context, adminIDs []string, status string) error {
	if len(adminIDs) == 0 {
		return nil
	}
	_, err := r.data.db.Admin.
		Update().
		Where(
			entadmin.IDIn(adminIDs...),
			entadmin.DeletedAtIsNil(),
		).
		SetStatus(entadmin.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// ExistsUsername 判断用户名是否已存在（excludeID 非空时排除该用户，仅统计未删除记录）
func (r *adminRepo) ExistsUsername(ctx context.Context, username, excludeID string) (bool, error) {
	return r.existsByField(ctx, entadmin.UsernameEQ(username), excludeID)
}

// ExistsPhone 判断手机号是否已存在（excludeID 非空时排除该用户）
func (r *adminRepo) ExistsPhone(ctx context.Context, phone, excludeID string) (bool, error) {
	return r.existsByField(ctx, entadmin.PhoneEQ(phone), excludeID)
}

// ExistsEmail 判断邮箱是否已存在（excludeID 非空时排除该用户）
func (r *adminRepo) ExistsEmail(ctx context.Context, email, excludeID string) (bool, error) {
	return r.existsByField(ctx, entadmin.EmailEQ(email), excludeID)
}

// existsByField 通用存在性查询：仅检查未软删除记录，excludeID 非空时排除自身
func (r *adminRepo) existsByField(ctx context.Context, p predicate.Admin, excludeID string) (bool, error) {
	query := r.data.db.Admin.Query().
		Where(p, entadmin.DeletedAtIsNil())
	if excludeID != "" {
		query = query.Where(entadmin.IDNEQ(excludeID))
	}
	return query.Exist(ctx)
}

// ====== 角色关联方法实现 ======

// getRoleIDsByAdmins 批量查询多个用户的角色ID（一次查询，按 adminID 分组）
func (r *adminRepo) getRoleIDsByAdmins(ctx context.Context, adminIDs []string) (map[string][]string, error) {
	if len(adminIDs) == 0 {
		return map[string][]string{}, nil
	}
	bindings, err := r.data.db.AdminRole.Query().
		Where(adminrole.AdminIDIn(adminIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string, len(adminIDs))
	for _, b := range bindings {
		m[b.AdminID] = append(m[b.AdminID], b.RoleID)
	}
	return m, nil
}

// getRoleNamesByIDs 根据角色ID列表批量查询角色名称（去重后一次查询）
func (r *adminRepo) getRoleNamesByIDs(ctx context.Context, roleIDs []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(roleIDs) == 0 {
		return result, nil
	}
	// 去重
	unique := make([]string, 0, len(roleIDs))
	seen := make(map[string]struct{}, len(roleIDs))
	for _, id := range roleIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	roles, err := r.data.db.Role.Query().
		Where(role.IDIn(unique...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range roles {
		result[v.ID] = v.Name
	}
	return result, nil
}

// getDepartmentNamesByIDs 根据部门ID列表批量查询部门名称（去重后一次查询，避免 N+1）
func (r *adminRepo) getDepartmentNamesByIDs(ctx context.Context, deptIDs []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(deptIDs) == 0 {
		return result, nil
	}
	// 去重
	unique := make([]string, 0, len(deptIDs))
	seen := make(map[string]struct{}, len(deptIDs))
	for _, id := range deptIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return result, nil
	}

	depts, err := r.data.db.Department.Query().
		Where(department.IDIn(unique...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range depts {
		result[v.ID] = v.Name
	}
	return result, nil
}

// GetDepartmentNamesByIDs 批量查询部门名称（个人中心归属部门展示用，复用去重查询逻辑）
func (r *adminRepo) GetDepartmentNamesByIDs(ctx context.Context, deptIDs []string) (map[string]string, error) {
	return r.getDepartmentNamesByIDs(ctx, deptIDs)
}

// fillDepartmentName 为单个用户填充归属部门名称（无部门或查询失败时静默处理）
func (r *adminRepo) fillDepartmentName(ctx context.Context, admin *biz.Admin) {
	if admin == nil || admin.DepartmentID == "" {
		return
	}
	m, err := r.getDepartmentNamesByIDs(ctx, []string{admin.DepartmentID})
	if err != nil {
		r.log.Warnf("获取用户归属部门名称失败: %v", err)
		return
	}
	admin.DepartmentName = m[admin.DepartmentID]
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

// ====== 数据权限规则关联方法实现 ======

// BindDataPermissionsForAdmin 为用户绑定数据权限规则（全量替换：先删除所有现有绑定，再批量插入新绑定）
func (r *adminRepo) BindDataPermissionsForAdmin(ctx context.Context, adminID string, dataPermissionIDs []string) error {
	// 开启事务
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	// 1. 删除该用户的所有现有绑定
	_, err = tx.AdminDataPermission.Delete().
		Where(entadmindataprm.AdminIDEQ(adminID)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		r.log.Errorf("删除用户 %s 的旧数据权限规则绑定失败: %v", adminID, err)
		return err
	}

	// 2. 批量插入新的绑定（如果有）
	if len(dataPermissionIDs) > 0 {
		builders := make([]*ent.AdminDataPermissionCreate, len(dataPermissionIDs))
		for i, dpID := range dataPermissionIDs {
			builders[i] = tx.AdminDataPermission.Create().
				SetID(uuid.GenerateXID()).
				SetAdminID(adminID).
				SetDataPermissionID(dpID)
		}

		err = tx.AdminDataPermission.CreateBulk(builders...).Exec(ctx)
		if err != nil {
			tx.Rollback()
			r.log.Errorf("为用户 %s 绑定数据权限规则失败: %v", adminID, err)
			return err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		r.log.Errorf("提交用户数据权限规则绑定事务失败: %v", err)
		return err
	}

	r.log.Infof("用户 %s 绑定了 %d 个数据权限规则", adminID, len(dataPermissionIDs))
	return nil
}

// GetDataPermissionIDsByAdmin 获取用户绑定的数据权限规则ID列表
func (r *adminRepo) GetDataPermissionIDsByAdmin(ctx context.Context, adminID string) ([]string, error) {
	bindings, err := r.data.db.AdminDataPermission.Query().
		Where(entadmindataprm.AdminIDEQ(adminID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(bindings))
	for _, b := range bindings {
		ids = append(ids, b.DataPermissionID)
	}
	return ids, nil
}

// GetDataPermissionsByAdminIDs 批量获取多个用户绑定的数据权限规则（返回 map[adminID][]*DataPermission）
func (r *adminRepo) GetDataPermissionsByAdminIDs(ctx context.Context, adminIDs []string) (map[string][]*biz.DataPermission, error) {
	result := make(map[string][]*biz.DataPermission)
	if len(adminIDs) == 0 {
		return result, nil
	}

	// 批量查询绑定关系
	bindings, err := r.data.db.AdminDataPermission.Query().
		Where(entadmindataprm.AdminIDIn(adminIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 收集全部规则ID，一次性查询规则详情，避免 N+1
	dpIDs := make([]string, 0, len(bindings))
	seen := make(map[string]struct{}, len(bindings))
	for _, b := range bindings {
		if _, ok := seen[b.DataPermissionID]; ok {
			continue
		}
		seen[b.DataPermissionID] = struct{}{}
		dpIDs = append(dpIDs, b.DataPermissionID)
	}

	dpByID := make(map[string]*ent.DataPermission)
	if len(dpIDs) > 0 {
		perms, err := r.data.db.DataPermission.Query().
			Where(datapermission.DeletedAtIsNil(), datapermission.StatusEQ("enabled")).
			Where(datapermission.IDIn(dpIDs...)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range perms {
			dpByID[p.ID] = p
		}
	}

	for _, b := range bindings {
		p, ok := dpByID[b.DataPermissionID]
		if !ok {
			continue
		}
		result[b.AdminID] = append(result[b.AdminID], &biz.DataPermission{
			ID:        p.ID,
			Name:      p.Name,
			Code:      p.Code,
			ScopeType: p.ScopeType,
		})
	}
	return result, nil
}

// getDataPermissionIDsByAdmins 批量查询多个用户绑定的数据权限规则ID（一次查询，按 adminID 分组，避免 N+1）
func (r *adminRepo) getDataPermissionIDsByAdmins(ctx context.Context, adminIDs []string) (map[string][]string, error) {
	if len(adminIDs) == 0 {
		return map[string][]string{}, nil
	}
	bindings, err := r.data.db.AdminDataPermission.Query().
		Where(entadmindataprm.AdminIDIn(adminIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string, len(adminIDs))
	for _, b := range bindings {
		m[b.AdminID] = append(m[b.AdminID], b.DataPermissionID)
	}
	return m, nil
}

// getDataPermissionNamesByIDs 根据数据权限规则ID列表批量查询规则名称（去重后一次查询，避免 N+1）
func (r *adminRepo) getDataPermissionNamesByIDs(ctx context.Context, dpIDs []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(dpIDs) == 0 {
		return result, nil
	}
	// 去重
	unique := make([]string, 0, len(dpIDs))
	seen := make(map[string]struct{}, len(dpIDs))
	for _, id := range dpIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return result, nil
	}

	perms, err := r.data.db.DataPermission.Query().
		Where(datapermission.IDIn(unique...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range perms {
		result[v.ID] = v.Name
	}
	return result, nil
}
