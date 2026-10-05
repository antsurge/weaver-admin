package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// FormSchema 表单构建器-表单定义表
// 保存可视化设计的表单 JSON Schema 定义，供"表单构建器"模块设计、预览与渲染。
type FormSchema struct {
	ent.Schema
}

func (FormSchema) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "form_schema"},
		entsql.WithComments(true),
		schema.Comment("表单构建器-表单定义表"),
	}
}

// Mixin of the FormSchema.
func (FormSchema) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (FormSchema) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("表单ID"),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("表单名称"),

		field.String("code").
			NotEmpty().
			MaxLen(64).
			Unique().
			Comment("表单编码（唯一，供运行时渲染引用）"),

		field.String("description").
			Optional().
			MaxLen(255).
			Comment("表单描述"),

		// 表单设计内容：JSON Schema（字段列表 + 布局配置）
		field.JSON("schema_json", map[string]any{}).
			Optional().
			Comment("表单 JSON Schema 定义"),

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

func (FormSchema) Edges() []ent.Edge {
	return nil
}
