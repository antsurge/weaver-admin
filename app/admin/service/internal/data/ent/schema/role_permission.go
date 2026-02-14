package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
)

type RolePermission struct {
	ent.Schema
}

func (RolePermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "role_permission"},
		entsql.WithComments(true),
		schema.Comment("角色和权限关联表"),
	}
}

func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable().
			MaxLen(36).
			Comment("主键ID"),
		field.String("role_id").
			MaxLen(36).
			Comment("角色ID"),
		field.String("permission_id").
			MaxLen(36).
			Comment("权限ID"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
	}
}

func (RolePermission) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	edge.From("role", Role.Type).
	//		Ref("permissions"). // 对应 Role.permissions
	//		Field("role_id").
	//		Unique(),
	//
	//	edge.From("permission", Permission.Type).
	//		Ref("roles"). // 对应 Permission.roles（需在 Permission 中定义）
	//		Field("permission_id").
	//		Unique(),
	//}
}
