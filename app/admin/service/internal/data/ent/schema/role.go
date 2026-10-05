package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Role struct {
	ent.Schema
}

// Annotations 指定表名
func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "role"},
		entsql.WithComments(true),
		schema.Comment("系统角色表"),
	}
}

// Mixin of the Role.
func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique(),
		field.String("name").
			Comment("角色名称"),
		field.String("code").
			Unique().
			Comment("角色编码"),
		field.Int("weight").
			Default(0).
			Comment("排序权重"),
		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),
		field.String("remark").
			Optional().
			Comment("备注"),
		field.Bool("is_system").
			Default(false).
			Comment("是否系统内置"),

		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		// Role → Menus (多对多，通过 role_menus 关联表)
		edge.To("menus", Menu.Type).
			Through("role_menus", RoleMenu.Type),

		// Role → DataPermissions (多对多，通过 role_data_permissions 关联表)
		edge.To("data_permissions", DataPermission.Type).
			Through("role_data_permissions", RoleDataPermission.Type),

		// Role ← Admins (反向，用户-角色多对多)
		edge.From("admins", Admin.Type).
			Ref("roles").
			Through("admin_roles", AdminRole.Type),
	}
}
