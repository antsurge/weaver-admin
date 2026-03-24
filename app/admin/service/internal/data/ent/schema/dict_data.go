package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type DictData struct {
	ent.Schema
}

func (DictData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dict_data"},
		entsql.WithComments(true),
		schema.Comment("字典数据表"),
	}
}

func (DictData) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("字典数据ID"),

		field.String("dict_type_id").
			NotEmpty().
			MaxLen(36).
			Comment("字典类型ID"),

		field.String("label").
			NotEmpty().
			MaxLen(64).
			Comment("标签"),

		field.String("value").
			NotEmpty().
			MaxLen(64).
			Comment("值"),

		field.Int("weight").
			Default(0).
			Comment("排序"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),

		field.JSON("extension", map[string]any{}).
			Optional().
			Comment("扩展参数"),

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

func (DictData) Edges() []ent.Edge {
	return nil
}
