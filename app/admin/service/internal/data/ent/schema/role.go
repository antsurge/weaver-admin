package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
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

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),
		field.String("name").NotEmpty().MaxLen(64).Comment("角色名称"),
		field.String("code").NotEmpty().Unique().MaxLen(64).Comment("角色唯一编码"),
		field.String("description").Optional().MaxLen(256).Comment("角色描述"),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Role) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	// Role → Permissions
	//	edge.To("permissions", Permission.Type).
	//		Through("role_permissions", RolePermission.Type),
	//
	//	// Role ← Admins (反向)
	//	edge.From("admins", Admin.Type).
	//		Ref("roles"), // 对应 Admin.roles
	//}
}
