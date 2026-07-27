package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Menu struct {
	ent.Schema
}

func (Menu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "menu"},
		entsql.WithComments(true),
		schema.Comment("菜单表"),
	}
}

func (Menu) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),
		field.String("parent_id").Default("").Comment("父菜单ID, null表示根节点"),
		field.String("name").NotEmpty().MaxLen(64).Comment("菜单名称"),
		// 前端表单已不再提交 code，放宽为可选非唯一（保留字段兼容存量数据与列表搜索）
		field.String("code").Default("").MaxLen(64).Comment("菜单唯一编码"),
		field.String("title").Default("").MaxLen(60).Comment("标题（国际化 key）"),
		field.String("remark").Optional().MaxLen(256).Comment("备注"),
		field.String("path").Default("").MaxLen(256).Comment("路由路径"),
		field.String("icon").Default("").MaxLen(128).Comment("图标"),
		// 枚举扩展：保留旧值 menu_dir/button 兼容存量数据，新增前端值 catalog/iframe/link/action
		field.Enum("type").Values("menu_dir", "catalog", "menu", "iframe", "link", "action", "button").Default("menu_dir").Comment("类型:catalog=菜单目录,menu=菜单项,iframe=内嵌,link=外链,action=页面按钮"),
		field.Enum("menu_type").Values("tab", "link", "iframe").Default("tab").Comment("菜单类型:tab=选项卡,link=链接(站外),iframe=Iframe"),
		field.String("url").Default("").MaxLen(256).Comment("链接地址（iframe/外链）"),
		field.String("component").Default("").MaxLen(256).Comment("组件路径"),
		field.String("auth_code").Default("").MaxLen(100).Comment("权限标识"),
		field.String("badge_type").Default("").MaxLen(20).Comment("徽标类型:dot=圆点,text=文本"),
		field.String("badge").Default("").MaxLen(10).Comment("徽标内容"),
		field.String("badge_variants").Default("").MaxLen(20).Comment("徽标样式:default/destructive/primary/success/warning"),
		field.Int("weight").Default(0).Comment("权重"),
		field.Enum("status").Values("enabled", "disabled").Default("enabled").Comment("状态：enabled=启用 disabled=禁用"),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Menu) Edges() []ent.Edge {
	return []ent.Edge{
		// Menu ← Roles (反向关联)
		edge.From("roles", Role.Type).
			Ref("menus").
			Through("role_menus", RoleMenu.Type),
	}
}
