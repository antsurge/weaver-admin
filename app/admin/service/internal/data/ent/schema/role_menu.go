package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RoleMenu struct {
	ent.Schema
}

// Annotations 指定表名
func (RoleMenu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "role_menus"},
		entsql.WithComments(true),
		schema.Comment("角色菜单关联表"),
	}
}

func (RoleMenu) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique().
			Comment("主键ID"),
		field.String("role_id").
			NotEmpty().
			Comment("角色ID"),
		field.String("menu_id").
			NotEmpty().
			Comment("菜单ID"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

func (RoleMenu) Edges() []ent.Edge {
	return []ent.Edge{
		// RoleMenu → Role
		edge.To("role", Role.Type).
			Unique().
			Field("role_id").
			Required(),
		// RoleMenu → Menu
		edge.To("menu", Menu.Type).
			Unique().
			Field("menu_id").
			Required(),
	}
}
