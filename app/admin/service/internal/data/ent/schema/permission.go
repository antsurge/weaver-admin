package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
)

type Permission struct {
	ent.Schema
}

func (Permission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "permission"},
		entsql.WithComments(true),
		schema.Comment("权限表"),
	}
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),
		field.String("name").NotEmpty().MaxLen(64).Comment("权限名称"),
		field.String("code").NotEmpty().Unique().MaxLen(64).Comment("权限唯一编码"),
		field.String("description").Optional().MaxLen(256).Comment("权限描述"),
		field.Int("parent_id").Optional().Default(0).Comment("父权限ID, 0表示根节点"),
		field.String("type").Default("api").MaxLen(32).Comment("类型: api/menu"),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Permission) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	// 反向：哪些角色拥有此权限
	//	edge.From("roles", Role.Type).
	//		Ref("permissions").
	//		Through("role_permissions", RolePermission.Type),
	//}
}
