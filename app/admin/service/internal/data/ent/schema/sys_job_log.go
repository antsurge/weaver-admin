package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type SysJobLog struct {
	ent.Schema
}

func (SysJobLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_job_log"},
		entsql.WithComments(true),
		schema.Comment("定时任务执行日志表"),
	}
}

// Mixin of the SysJobLog.
func (SysJobLog) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (SysJobLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("日志ID"),

		field.String("job_id").
			NotEmpty().
			MaxLen(36).
			Comment("任务ID"),

		field.String("job_name").
			Optional().
			Default("").
			MaxLen(64).
			Comment("任务名称"),

		field.String("job_group").
			Optional().
			Default("").
			MaxLen(64).
			Comment("任务分组"),

		field.String("invoke_target").
			Optional().
			Default("").
			MaxLen(512).
			Comment("调用目标"),

		field.String("cron_expression").
			Optional().
			Default("").
			MaxLen(64).
			Comment("Cron 表达式"),

		field.Enum("status").
			Values("success", "fail").
			Default("success").
			Comment("执行结果"),

		field.String("job_message").
			Optional().
			Default("").
			MaxLen(2000).
			Comment("执行信息"),

		field.Time("start_time").
			Optional().
			Nillable().
			Comment("开始时间"),

		field.Time("end_time").
			Optional().
			Nillable().
			Comment("结束时间"),

		field.Int64("duration").
			Default(0).
			Comment("耗时（毫秒）"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
	}
}

func (SysJobLog) Edges() []ent.Edge {
	return nil
}
