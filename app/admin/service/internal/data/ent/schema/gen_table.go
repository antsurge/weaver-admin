package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// GenTable 代码生成器-生成表配置表
// 记录代码生成的来源表、模块/业务命名与字段配置，供代码生成器按模板产出前后端代码。
type GenTable struct {
	ent.Schema
}

func (GenTable) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gen_table"},
		entsql.WithComments(true),
		schema.Comment("代码生成器-生成表配置表"),
	}
}

// Mixin of the GenTable.
func (GenTable) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (GenTable) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("配置ID"),

		// 数据库表名（信息源，如 sys_job）
		field.String("table_name").
			NotEmpty().
			MaxLen(100).
			Comment("数据库表名"),

		field.String("table_comment").
			Optional().
			MaxLen(255).
			Comment("表备注"),

		// 生成代码所属模块（如 system / permission / organization）
		field.String("module_name").
			NotEmpty().
			MaxLen(50).
			Comment("模块名"),

		// 业务名（驼峰，如 Job / JobLog）
		field.String("biz_name").
			NotEmpty().
			MaxLen(100).
			Comment("业务名（驼峰）"),

		// 生成的字段配置（JSON 数组：列名/类型/注释/是否列表展示等）
		field.JSON("fields_json", map[string]any{}).
			Optional().
			Comment("字段配置"),

		// 生成类型：single=单表 CRUD；（预留 tree=树表，可扩展）
		field.String("gen_type").
			Optional().
			MaxLen(20).
			Comment("生成类型"),

		// 是否生成菜单（落库菜单树）
		field.Bool("menu_enabled").
			Default(false).
			Comment("是否生成菜单"),

		// 菜单所属模块（父目录归属，为空则用模块名）
		field.String("menu_module").
			Optional().
			MaxLen(50).
			Comment("菜单所属模块"),

		// 生成的按钮列表（list/detail/create/edit/delete/batchDelete/status/import/export）
		field.JSON("buttons_json", []string{}).
			Optional().
			Comment("生成的按钮列表"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),

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

func (GenTable) Edges() []ent.Edge {
	return nil
}
