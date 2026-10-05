package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type DataPermission struct {
	ent.Schema
}

func (DataPermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "data_permission"},
		entsql.WithComments(true),
		schema.Comment("数据权限规则表"),
	}
}

// Mixin of the DataPermission.
func (DataPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (DataPermission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("规则ID"),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("规则名称"),

		field.String("code").
			NotEmpty().
			MaxLen(100).
			Unique().
			Comment("规则编码"),

		field.String("scope_type").
			NotEmpty().
			MaxLen(1).
			Default("1").
			Comment("数据范围（1=全部数据 2=自定义 3=本部门 4=本部门及以下 5=仅本人）"),

		field.String("dept_ids").
			Optional().
			Default("[]").
			MaxLen(2000).
			Comment("自定义数据范围的部门ID集合（JSON数组）"),

		field.String("role_ids").
			Optional().
			Default("[]").
			MaxLen(2000).
			Comment("自定义数据范围的角色ID集合（JSON数组）"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),

		field.String("remark").
			Optional().
			MaxLen(256).
			Comment("备注"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
	}
}

func (DataPermission) Edges() []ent.Edge {
	return []ent.Edge{
		// DataPermission ← Roles (反向，角色-数据权限规则多对多)
		edge.From("roles", Role.Type).
			Ref("data_permissions").
			Through("role_data_permissions", RoleDataPermission.Type),
		// DataPermission ← Admins (反向，管理员-数据权限规则多对多)
		edge.From("admins", Admin.Type).
			Ref("data_permissions").
			Through("admin_data_permissions", AdminDataPermission.Type),
	}
}
