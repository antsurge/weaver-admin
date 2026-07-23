package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type AdminRole struct {
	ent.Schema
}

func (AdminRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "admin_roles"},
		entsql.WithComments(true),
		schema.Comment("用户角色关联表"),
	}
}

func (AdminRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique().
			Comment("主键ID"),
		field.String("admin_id").
			NotEmpty().
			Comment("用户ID"),
		field.String("role_id").
			NotEmpty().
			Comment("角色ID"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

func (AdminRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("admin", Admin.Type).
			Unique().
			Field("admin_id").
			Required(),
		edge.To("role", Role.Type).
			Unique().
			Field("role_id").
			Required(),
	}
}
