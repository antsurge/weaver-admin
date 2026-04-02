package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Position struct {
	ent.Schema
}

func (Position) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "position"},
		entsql.WithComments(true),
		schema.Comment("职务表"),
	}
}

func (Position) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable().
			MaxLen(36).
			Comment("ID"),
		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("职务名称"),
		field.String("code").
			NotEmpty().
			MaxLen(64).
			Comment("职务编码"),
		field.Int("weight").
			Default(0).
			Comment("权重/职务级别（越小职务越高）"),
		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态：enabled=启用 disabled=禁用"),
		field.String("remark").
			Optional().
			MaxLen(256).
			Comment("备注"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
		field.String("created_by").
			Immutable().
			Default("").
			Comment("创建人"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
		field.String("updated_by").
			Default("").
			Comment("更新人"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
		field.String("deleted_by").
			Optional().
			Nillable().
			Comment("删除人"),
	}
}

func (Position) Edges() []ent.Edge {
	return nil
}

func (Position) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code", "deleted_at").Unique(),
	}
}
