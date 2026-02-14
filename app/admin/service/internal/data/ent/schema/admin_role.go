package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
)

type AdminRole struct {
	ent.Schema
}

func (AdminRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "admin_role"},
		entsql.WithComments(true),
		schema.Comment("用户和角色关联表"),
	}
}

func (AdminRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable().
			MaxLen(36).
			Comment("主键ID"),
		field.String("admin_id").
			MaxLen(36).
			Comment("用户ID"),
		field.String("role_id").
			MaxLen(36).
			Comment("角色ID"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
	}
}

func (AdminRole) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	// 从 admin_id → Admin
	//	edge.From("admin", Admin.Type).
	//		Ref("roles"). // 对应 Admin.roles
	//		Field("admin_id").
	//		Unique(),
	//
	//	// 从 role_id → Role
	//	edge.From("role", Role.Type).
	//		Ref("admins"). // 对应 Role.admins
	//		Field("role_id").
	//		Unique(),
	//}
}
