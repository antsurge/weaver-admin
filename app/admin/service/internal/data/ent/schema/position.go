package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Position struct {
	ent.Schema
}

func (Position) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "position"},
		entsql.WithComments(true),
		schema.Comment("岗位表"),
	}
}

func (Position) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("岗位名称"),

		field.String("code").
			NotEmpty().
			Unique().
			MaxLen(64).
			Comment("岗位code"),

		field.Int("weight").
			Default(0).
			Comment("权重"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态：enabled=启用 disabled=禁用"),

		field.String("description").Optional().MaxLen(256).Comment("描述"),

		field.Time("created_at").
			Immutable().
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Position) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	// 反向：哪些角色拥有此权限
	//	edge.From("roles", Role.Type).
	//		Ref("permissions").
	//		Through("role_permissions", RolePermission.Type),
	//}
}
