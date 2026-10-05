package data

import (
	"context"
	"errors"

	entgo "entgo.io/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/admin"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/adminrole"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/datapermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/department"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/dictdata"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/dicttype"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/formschema"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/gentable"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/loginlog"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menu"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menuapipermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/notification"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/notificationrecord"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/operationlog"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/position"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/role"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/rolemenu"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/rolepermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysconfig"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysjob"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysjoblog"
	"github.com/antsurge/weaver-admin/pkg/tenant"
)

// ErrTenantRequired 表示当前操作缺少租户作用域，禁止执行。
var ErrTenantRequired = errors.New("缺少租户作用域，拒绝执行")

// tenantField 所有接入 TenantMixin 的实体共有的租户字段名。
const tenantField = "tenant_id"

// TenantQueryInterceptor 查询拦截器：
//   - 上下文中携带租户ID时，自动追加 tenant_id = ? 过滤条件；
//   - 平台级（tenant.IsUnscoped）或缺少租户上下文时，不追加过滤。
//
// 保证业务代码无需感知租户，查询天然隔离。
func TenantQueryInterceptor() entgo.Interceptor {
	return entgo.InterceptFunc(func(next entgo.Querier) entgo.Querier {
		return entgo.QuerierFunc(func(ctx context.Context, query entgo.Query) (entgo.Value, error) {
			tid, ok := tenant.From(ctx)
			if ok && !tenant.IsUnscoped(ctx) && tid != "" {
				filterByTenant(query, tid)
			}
			return next.Query(ctx, query)
		})
	})
}

// filterByTenant 为查询追加租户过滤条件。
func filterByTenant(query entgo.Query, tid string) {
	switch q := query.(type) {
	case *ent.AdminQuery:
		q.Where(admin.TenantIDEQ(tid))
	case *ent.AdminRoleQuery:
		q.Where(adminrole.TenantIDEQ(tid))
	case *ent.DataPermissionQuery:
		q.Where(datapermission.TenantIDEQ(tid))
	case *ent.DepartmentQuery:
		q.Where(department.TenantIDEQ(tid))
	case *ent.DictDataQuery:
		q.Where(dictdata.TenantIDEQ(tid))
	case *ent.DictTypeQuery:
		q.Where(dicttype.TenantIDEQ(tid))
	case *ent.FormSchemaQuery:
		q.Where(formschema.TenantIDEQ(tid))
	case *ent.GenTableQuery:
		q.Where(gentable.TenantIDEQ(tid))
	case *ent.LoginLogQuery:
		q.Where(loginlog.TenantIDEQ(tid))
	case *ent.MenuQuery:
		q.Where(menu.TenantIDEQ(tid))
	case *ent.MenuApiPermissionQuery:
		q.Where(menuapipermission.TenantIDEQ(tid))
	case *ent.NotificationQuery:
		q.Where(notification.TenantIDEQ(tid))
	case *ent.NotificationRecordQuery:
		q.Where(notificationrecord.TenantIDEQ(tid))
	case *ent.OperationLogQuery:
		q.Where(operationlog.TenantIDEQ(tid))
	case *ent.PositionQuery:
		q.Where(position.TenantIDEQ(tid))
	case *ent.RoleQuery:
		q.Where(role.TenantIDEQ(tid))
	case *ent.RoleMenuQuery:
		q.Where(rolemenu.TenantIDEQ(tid))
	case *ent.RolePermissionQuery:
		q.Where(rolepermission.TenantIDEQ(tid))
	case *ent.SysConfigQuery:
		q.Where(sysconfig.TenantIDEQ(tid))
	case *ent.SysJobQuery:
		q.Where(sysjob.TenantIDEQ(tid))
	case *ent.SysJobLogQuery:
		q.Where(sysjoblog.TenantIDEQ(tid))
	}
}

// TenantMutationHook 变更钩子：
//   - 创建（OpCreate）：强制将 tenant_id 填充为当前租户，杜绝串租户写入；
//   - 更新/删除：追加 tenant_id 过滤，仅允许操作当前租户数据；
//   - 平台级（Unscoped）或该实体无租户字段时，原样放行。
//
// 写操作若拿不到租户作用域直接报错，防止静默写入错误租户。
func TenantMutationHook() entgo.Hook {
	return func(next entgo.Mutator) entgo.Mutator {
		return entgo.MutateFunc(func(ctx context.Context, m entgo.Mutation) (entgo.Value, error) {
			// 该实体无租户字段（平台级表）直接放行。
			if _, ok := m.Field(tenantField); !ok {
				return next.Mutate(ctx, m)
			}
			// 平台级操作（Unscoped）放行。
			if tenant.IsUnscoped(ctx) {
				return next.Mutate(ctx, m)
			}
			tid, ok := tenant.From(ctx)
			if !ok || tid == "" {
				return nil, ErrTenantRequired
			}
			switch op := m.Op(); {
			case op.Is(entgo.OpCreate):
				if err := m.SetField(tenantField, tid); err != nil {
					return nil, err
				}
			case op.Is(entgo.OpUpdate) || op.Is(entgo.OpUpdateOne) ||
				op.Is(entgo.OpDelete) || op.Is(entgo.OpDeleteOne):
				filterMutationByTenant(m, tid)
			}
			return next.Mutate(ctx, m)
		})
	}
}

// filterMutationByTenant 为更新/删除变更追加租户过滤条件。
func filterMutationByTenant(m entgo.Mutation, tid string) {
	switch mm := m.(type) {
	case *ent.AdminMutation:
		mm.Where(admin.TenantIDEQ(tid))
	case *ent.AdminRoleMutation:
		mm.Where(adminrole.TenantIDEQ(tid))
	case *ent.DataPermissionMutation:
		mm.Where(datapermission.TenantIDEQ(tid))
	case *ent.DepartmentMutation:
		mm.Where(department.TenantIDEQ(tid))
	case *ent.DictDataMutation:
		mm.Where(dictdata.TenantIDEQ(tid))
	case *ent.DictTypeMutation:
		mm.Where(dicttype.TenantIDEQ(tid))
	case *ent.FormSchemaMutation:
		mm.Where(formschema.TenantIDEQ(tid))
	case *ent.GenTableMutation:
		mm.Where(gentable.TenantIDEQ(tid))
	case *ent.LoginLogMutation:
		mm.Where(loginlog.TenantIDEQ(tid))
	case *ent.MenuMutation:
		mm.Where(menu.TenantIDEQ(tid))
	case *ent.MenuApiPermissionMutation:
		mm.Where(menuapipermission.TenantIDEQ(tid))
	case *ent.NotificationMutation:
		mm.Where(notification.TenantIDEQ(tid))
	case *ent.NotificationRecordMutation:
		mm.Where(notificationrecord.TenantIDEQ(tid))
	case *ent.OperationLogMutation:
		mm.Where(operationlog.TenantIDEQ(tid))
	case *ent.PositionMutation:
		mm.Where(position.TenantIDEQ(tid))
	case *ent.RoleMutation:
		mm.Where(role.TenantIDEQ(tid))
	case *ent.RoleMenuMutation:
		mm.Where(rolemenu.TenantIDEQ(tid))
	case *ent.RolePermissionMutation:
		mm.Where(rolepermission.TenantIDEQ(tid))
	case *ent.SysConfigMutation:
		mm.Where(sysconfig.TenantIDEQ(tid))
	case *ent.SysJobMutation:
		mm.Where(sysjob.TenantIDEQ(tid))
	case *ent.SysJobLogMutation:
		mm.Where(sysjoblog.TenantIDEQ(tid))
	}
}
