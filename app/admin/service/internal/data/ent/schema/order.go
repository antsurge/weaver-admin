package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Order 订单
type Order struct {
	ent.Schema
}

func (Order) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "order"},
		entsql.WithComments(true),
		schema.Comment("订单"),
	}
}

// Mixin of the Order.
func (Order) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("ID"),
		field.String("name").
			Optional().
			Comment("name"),
		field.String("sn").
			Optional().
			Comment("sn"),
		field.Enum("status").Values("enabled", "disabled").Default("enabled").Comment("状态：enabled=启用 disabled=禁用"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
	}
}
