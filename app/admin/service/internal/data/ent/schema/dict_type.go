package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type DictType struct {
	ent.Schema
}

func (DictType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dict_type"},
		entsql.WithComments(true),
		schema.Comment("字典类型表"),
	}
}

func (DictType) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("字典ID"),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("字典名称"),

		field.String("code").
			NotEmpty().
			MaxLen(64).
			Unique().
			Comment("字典编码"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),

		field.String("remark").
			Optional().
			MaxLen(256).
			Comment("备注"),

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

func (DictType) Edges() []ent.Edge {
	return nil
}
