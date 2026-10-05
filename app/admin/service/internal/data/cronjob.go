package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysjob"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysjoblog"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type cronJobRepo struct {
	data *Data
	log  *log.Helper
}

func NewCronJobRepo(data *Data, logger log.Logger) biz.CronJobRepo {
	return &cronJobRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// redisJobLocker 基于 redis SetNX 的分布式锁
type redisJobLocker struct {
	rdb *redis.Client
}

// NewRedisJobLocker 创建分布式锁
func NewRedisJobLocker(data *Data) biz.JobLocker {
	return &redisJobLocker{rdb: data.GetRedis()}
}

func (l *redisJobLocker) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, func(), error) {
	ok, err := l.rdb.SetNX(ctx, key, time.Now().Unix(), ttl).Result()
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}
	unlock := func() {
		// 仅当锁属于自己时删除（使用 Lua 保证原子性）
		l.rdb.Eval(ctx,
			"if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end",
			[]string{key}, time.Now().Unix(),
		)
	}
	return true, unlock, nil
}

func (r *cronJobRepo) ListJob(ctx context.Context, req *biz.ListJobRequest) (*biz.ListJobResponse, error) {
	query := r.data.db.SysJob.Query().
		Where(sysjob.DeletedAtIsNil())

	if v := req.Name; len(v) > 0 {
		query = query.Where(sysjob.NameContains(v))
	}
	if v := req.JobGroup; len(v) > 0 {
		query = query.Where(sysjob.JobGroupContains(v))
	}
	if v := req.Status; len(v) > 0 {
		query = query.Where(sysjob.StatusEQ(sysjob.Status(v)))
	}

	params := enthelper.PaginationParams{
		CurrentPage: req.CurrentPage,
		PageSize:    req.PageSize,
	}
	params.Normalize()

	result, err := enthelper.Pagination(ctx, query.Order(ent.Desc(sysjob.FieldUpdatedAt)), params)
	if err != nil {
		return nil, err
	}

	list := make([]*biz.SysJob, 0, len(result.Data))
	for _, v := range result.Data {
		list = append(list, toBizJob(v))
	}
	return &biz.ListJobResponse{Data: list, Total: result.Total}, nil
}

func (r *cronJobRepo) GetJobByID(ctx context.Context, id string) (*biz.SysJob, error) {
	v, err := r.data.db.SysJob.Query().
		Where(sysjob.ID(id), sysjob.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toBizJob(v), nil
}

func (r *cronJobRepo) ListEnabledJobs(ctx context.Context) ([]*biz.SysJob, error) {
	list, err := r.data.db.SysJob.Query().
		Where(sysjob.StatusEQ(sysjob.Status(biz.JobStatusEnabled)), sysjob.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*biz.SysJob, 0, len(list))
	for _, v := range list {
		res = append(res, toBizJob(v))
	}
	return res, nil
}

func (r *cronJobRepo) CreateJob(ctx context.Context, p *biz.SysJob) error {
	_, err := r.data.db.SysJob.Create().
		SetID(p.ID).
		SetName(p.Name).
		SetJobGroup(p.JobGroup).
		SetJobType(p.JobType).
		SetInvokeTarget(p.InvokeTarget).
		SetCronExpression(p.CronExpression).
		SetMisfirePolicy(p.MisfirePolicy).
		SetConcurrent(p.Concurrent).
		SetStatus(sysjob.Status(p.Status)).
		SetNillableRemark(&p.Remark).
		SetCreatedAt(p.CreatedAt).
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx)
	return err
}

func (r *cronJobRepo) UpdateJob(ctx context.Context, p *biz.SysJob) error {
	_, err := r.data.db.SysJob.UpdateOneID(p.ID).
		SetName(p.Name).
		SetJobGroup(p.JobGroup).
		SetJobType(p.JobType).
		SetInvokeTarget(p.InvokeTarget).
		SetCronExpression(p.CronExpression).
		SetMisfirePolicy(p.MisfirePolicy).
		SetConcurrent(p.Concurrent).
		SetStatus(sysjob.Status(p.Status)).
		SetNillableRemark(&p.Remark).
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx)
	return err
}

func (r *cronJobRepo) UpdateJobStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.SysJob.UpdateOneID(id).
		SetStatus(sysjob.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *cronJobRepo) UpdateNextRunTime(ctx context.Context, id string, t *time.Time) error {
	// 由后台调度器在注册任务时调用，ctx 无租户上下文，使用 Unscoped 跳过拦截器。
	_, err := r.data.db.SysJob.UpdateOneID(id).
		SetNillableNextRunTime(t).
		SetUpdatedAt(time.Now()).
		Save(tenant.Unscoped(ctx))
	return err
}

func (r *cronJobRepo) DeleteJob(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	_, err := r.data.db.SysJob.Update().
		Where(sysjob.IDIn(ids...), sysjob.DeletedAtIsNil()).
		SetDeletedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	return err
}

func (r *cronJobRepo) ListJobLog(ctx context.Context, req *biz.ListJobLogRequest) (*biz.ListJobLogResponse, error) {
	query := r.data.db.SysJobLog.Query().
		Where(sysjoblog.DeletedAtIsNil())

	if v := req.JobID; len(v) > 0 {
		query = query.Where(sysjoblog.JobID(v))
	}
	if v := req.JobName; len(v) > 0 {
		query = query.Where(sysjoblog.JobNameContains(v))
	}
	if v := req.Status; len(v) > 0 {
		query = query.Where(sysjoblog.StatusEQ(sysjoblog.Status(v)))
	}

	params := enthelper.PaginationParams{
		CurrentPage: req.CurrentPage,
		PageSize:    req.PageSize,
	}
	params.Normalize()

	result, err := enthelper.Pagination(ctx, query.Order(ent.Desc(sysjoblog.FieldCreatedAt)), params)
	if err != nil {
		return nil, err
	}

	list := make([]*biz.SysJobLog, 0, len(result.Data))
	for _, v := range result.Data {
		list = append(list, toBizJobLog(v))
	}
	return &biz.ListJobLogResponse{Data: list, Total: result.Total}, nil
}

func (r *cronJobRepo) GetJobLogByID(ctx context.Context, id string) (*biz.SysJobLog, error) {
	v, err := r.data.db.SysJobLog.Query().
		Where(sysjoblog.ID(id), sysjoblog.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toBizJobLog(v), nil
}

func (r *cronJobRepo) CreateJobLog(ctx context.Context, p *biz.SysJobLog) error {
	if p.TenantID == "" {
		p.TenantID = tenant.DefaultTenantID
	}
	builder := r.data.db.SysJobLog.Create().
		SetID(p.ID).
		SetTenantID(p.TenantID).
		SetJobID(p.JobID).
		SetJobName(p.JobName).
		SetJobGroup(p.JobGroup).
		SetInvokeTarget(p.InvokeTarget).
		SetCronExpression(p.CronExpression).
		SetStatus(sysjoblog.Status(p.Status)).
		SetJobMessage(p.JobMessage).
		SetDuration(p.Duration).
		SetCreatedAt(p.CreatedAt)

	if p.StartTime != nil {
		builder = builder.SetStartTime(*p.StartTime)
	}
	if p.EndTime != nil {
		builder = builder.SetEndTime(*p.EndTime)
	}
	// 由后台调度器写入执行日志，ctx 无租户上下文，
	// 租户已在 p.TenantID 中显式指定，此处使用 Unscoped 跳过拦截器填充。
	_, err := builder.Save(tenant.Unscoped(ctx))
	return err
}

func (r *cronJobRepo) DeleteJobLog(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.data.db.SysJobLog.Update().
		Where(sysjoblog.IDIn(ids...), sysjoblog.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *cronJobRepo) ClearJobLog(ctx context.Context) error {
	_, err := r.data.db.SysJobLog.Update().
		Where(sysjoblog.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	return err
}

func toBizJob(v *ent.SysJob) *biz.SysJob {
	return &biz.SysJob{
		ID:             v.ID,
		TenantID:       v.TenantID,
		Name:           v.Name,
		JobGroup:       v.JobGroup,
		JobType:        v.JobType,
		InvokeTarget:   v.InvokeTarget,
		CronExpression: v.CronExpression,
		MisfirePolicy:  v.MisfirePolicy,
		Concurrent:     v.Concurrent,
		Status:         string(v.Status),
		Remark:         v.Remark,
		NextRunTime:    v.NextRunTime,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
		DeletedAt:      v.DeletedAt,
	}
}

func toBizJobLog(v *ent.SysJobLog) *biz.SysJobLog {
	return &biz.SysJobLog{
		ID:             v.ID,
		TenantID:       v.TenantID,
		JobID:          v.JobID,
		JobName:        v.JobName,
		JobGroup:       v.JobGroup,
		InvokeTarget:   v.InvokeTarget,
		CronExpression: v.CronExpression,
		Status:         string(v.Status),
		JobMessage:     v.JobMessage,
		StartTime:      v.StartTime,
		EndTime:        v.EndTime,
		Duration:       v.Duration,
		CreatedAt:      v.CreatedAt,
		DeletedAt:      v.DeletedAt,
	}
}
