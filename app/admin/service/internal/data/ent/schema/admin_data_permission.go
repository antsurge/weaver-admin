package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type AdminDataPermission struct {
	ent.Schema
}

// Annotations 指定表名
func (AdminDataPermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "admin_data_permission"},
		entsql.WithComments(true),
		schema.Comment("管理员数据权限规则关联表"),
	}
}

// Mixin of the AdminDataPermission.
func (AdminDataPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (AdminDataPermission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique().
			Comment("主键ID"),
		field.String("admin_id").
			NotEmpty().
			Comment("管理员ID"),
		field.String("data_permission_id").
			NotEmpty().
			Comment("数据权限规则ID"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

func (AdminDataPermission) Edges() []ent.Edge {
	return []ent.Edge{
		// AdminDataPermission → Admin
		edge.To("admin", Admin.Type).
			Unique().
			Field("admin_id").
			Required(),
		// AdminDataPermission → DataPermission
		edge.To("data_permission", DataPermission.Type).
			Unique().
			Field("data_permission_id").
			Required(),
	}
}
