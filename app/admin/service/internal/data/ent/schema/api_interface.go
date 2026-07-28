package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ApiInterface 从 openapi.yaml 导入的接口信息
type ApiInterface struct {
	ent.Schema
}

func (ApiInterface) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "api_interface"},
		entsql.WithComments(true),
		schema.Comment("API 接口表（从 openapi.yaml 导入）"),
	}
}

func (ApiInterface) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),
		field.String("service").NotEmpty().MaxLen(128).Comment("服务全限定名，如 admin.service.v1.PermissionService"),
		field.String("tag").Default("").MaxLen(64).Comment("OpenAPI 标签，如 Permission"),
		field.String("method").NotEmpty().MaxLen(10).Comment("HTTP method: GET/POST/PUT/DELETE"),
		field.String("path").NotEmpty().MaxLen(256).Comment("接口路径，如 /admin/v1/menu"),
		field.String("summary").Default("").MaxLen(256).Comment("接口描述"),
		// 业务唯一键：service|method|path（去重用）
		field.String("code").Unique().NotEmpty().MaxLen(512).Comment("业务唯一键 service|method|path"),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}