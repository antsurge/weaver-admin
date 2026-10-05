package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FormSubmission 表单构建器-提交数据表
// 存储动态表单（表单构建器设计的表单）的用户提交数据，以 JSON 保存表单内容。
type FormSubmission struct {
	ent.Schema
}

func (FormSubmission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "form_submission"},
		entsql.WithComments(true),
		schema.Comment("表单构建器-提交数据表"),
	}
}

// Mixin of the FormSubmission.
func (FormSubmission) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (FormSubmission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("提交记录ID"),

		// 所属表单编码（关联 form_schema.code）
		field.String("form_code").
			NotEmpty().
			MaxLen(64).
			Comment("表单编码"),

		field.String("form_name").
			Optional().
			MaxLen(64).
			Comment("表单名称（冗余，便于列表展示）"),

		// 提交数据：key 为字段 key，value 为字段值
		field.JSON("submit_data", map[string]any{}).
			Optional().
			Comment("提交数据（JSON）"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("提交时间"),

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

func (FormSubmission) Edges() []ent.Edge {
	return nil
}

func (FormSubmission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("form_code"),
		index.Fields("created_at"),
	}
}
