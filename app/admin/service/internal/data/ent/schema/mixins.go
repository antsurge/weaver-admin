package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// TenantMixin 为业务表统一注入租户字段（tenant_id）与索引。
//
// 约定：所有租户级业务表都必须接入该 Mixin；平台级表
// （tenant / api_interface / api_permission / security_policy）不接入。
type TenantMixin struct {
	mixin.Schema
}

// Fields 注入 tenant_id（默认 default，创建后不可修改）。
func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant_id").
			MaxLen(36).
			Default("default").
			Immutable().
			Comment("租户ID"),
	}
}

// Indexes 为 tenant_id 建索引，保证按租户过滤的查询性能。
func (TenantMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
	}
}
