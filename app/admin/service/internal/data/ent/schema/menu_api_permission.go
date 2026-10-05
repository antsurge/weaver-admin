package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MenuApiPermission 菜单按钮与接口的绑定关系（menu_id + api_code）。
// 接口以 api_interface.code（service|METHOD|path）为稳定键关联：
// 接口重新导入（全量清空重建）后 id 会变化，但 code 保持不变，绑定依然有效；
// 接口信息（service/method/path/summary/tag）实时从 api_interface 反查，不依赖快照。
type MenuApiPermission struct {
	ent.Schema
}

func (MenuApiPermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "menu_api_permission"},
		entsql.WithComments(true),
		schema.Comment("菜单接口绑定表（菜单按钮 ↔ 接口 code）"),
	}
}

func (MenuApiPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (MenuApiPermission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),
		field.String("menu_id").NotEmpty().MaxLen(36).Comment("菜单ID"),
		field.String("api_code").NotEmpty().MaxLen(512).Comment("接口业务唯一键 service|METHOD|path（对应 api_interface.code）"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("创建时间"),
	}
}

func (MenuApiPermission) Indexes() []ent.Index {
	return []ent.Index{
		// 同一菜单下同一接口只允许绑定一次
		index.Fields("menu_id", "api_code").Unique(),
		// 按接口 code 反查（如导入后查找受影响菜单、对账使用）
		index.Fields("api_code"),
	}
}
