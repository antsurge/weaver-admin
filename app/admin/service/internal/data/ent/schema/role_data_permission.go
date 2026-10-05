package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RoleDataPermission struct {
	ent.Schema
}

// Annotations 指定表名
func (RoleDataPermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "role_data_permission"},
		entsql.WithComments(true),
		schema.Comment("角色数据权限规则关联表"),
	}
}

// Mixin of the RoleDataPermission.
func (RoleDataPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (RoleDataPermission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique().
			Comment("主键ID"),
		field.String("role_id").
			NotEmpty().
			Comment("角色ID"),
		field.String("data_permission_id").
			NotEmpty().
			Comment("数据权限规则ID"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

func (RoleDataPermission) Edges() []ent.Edge {
	return []ent.Edge{
		// RoleDataPermission → Role
		edge.To("role", Role.Type).
			Unique().
			Field("role_id").
			Required(),
		// RoleDataPermission → DataPermission
		edge.To("data_permission", DataPermission.Type).
			Unique().
			Field("data_permission_id").
			Required(),
	}
}
