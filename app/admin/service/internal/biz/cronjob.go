package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// 任务状态
const (
	JobStatusEnabled  = "enabled"
	JobStatusDisabled = "disabled"
)

// 错失执行策略
const (
	MisfirePolicyImmediately = "immediately" // 立即执行
	MisfirePolicyOnce        = "once"        // 执行一次
	MisfirePolicyIgnore      = "ignore"      // 放弃执行
)

// 任务类型
const (
	JobTypeHTTP = "http" // HTTP 调用
)

// SysJob 定时任务领域模型
type SysJob struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenantId,omitempty"`
	Name           string     `json:"name"`
	JobGroup       string     `json:"jobGroup"`
	JobType        string     `json:"jobType"`
	InvokeTarget   string     `json:"invokeTarget"`
	CronExpression string     `json:"cronExpression"`
	MisfirePolicy  string     `json:"misfirePolicy"`
	Concurrent     bool       `json:"concurrent"`
	Status         string     `json:"status"`
	Remark         string     `json:"remark"`
	NextRunTime    *time.Time `json:"nextRunTime"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

// SysJobLog 定时任务执行日志领域模型
type SysJobLog struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenantId,omitempty"`
	JobID          string     `json:"jobId"`
	JobName        string     `json:"jobName"`
	JobGroup       string     `json:"jobGroup"`
	InvokeTarget   string     `json:"invokeTarget"`
	CronExpression string     `json:"cronExpression"`
	Status         string     `json:"status"`
	JobMessage     string     `json:"jobMessage"`
	StartTime      *time.Time `json:"startTime"`
	EndTime        *time.Time `json:"endTime"`
	Duration       int64      `json:"duration"`
	CreatedAt      time.Time  `json:"createdAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

type ListJobRequest struct {
	enthelper.PaginationParams
	Name     string `form:"name" query:"name"`
	JobGroup string `form:"jobGroup" query:"jobGroup"`
	Status   string `form:"status" query:"status"`
}

type ListJobResponse struct {
	Data  []*SysJob
	Total int
}

type ListJobLogRequest struct {
	enthelper.PaginationParams
	JobID   string `form:"jobId" query:"jobId"`
	JobName string `form:"jobName" query:"jobName"`
	Status  string `form:"status" query:"status"`
}

type ListJobLogResponse struct {
	Data  []*SysJobLog
	Total int
}

// CronJobRepo 定时任务仓储接口
type CronJobRepo interface {
	// ListJob 任务分页列表
	ListJob(context.Context, *ListJobRequest) (*ListJobResponse, error)
	// GetJobByID 任务详情
	GetJobByID(context.Context, string) (*SysJob, error)
	// ListEnabledJobs 查询所有启用中的任务（调度器启动时加载）
	ListEnabledJobs(context.Context) ([]*SysJob, error)
	// CreateJob 创建任务
	CreateJob(context.Context, *SysJob) error
	// UpdateJob 修改任务
	UpdateJob(context.Context, *SysJob) error
	// UpdateJobStatus 更新任务状态
	UpdateJobStatus(context.Context, string, string) error
	// UpdateNextRunTime 更新下次执行时间
	UpdateNextRunTime(context.Context, string, *time.Time) error
	// DeleteJob 删除任务（批量，软删除）
	DeleteJob(context.Context, []string) error
	// ListJobLog 日志分页列表
	ListJobLog(context.Context, *ListJobLogRequest) (*ListJobLogResponse, error)
	// GetJobLogByID 日志详情
	GetJobLogByID(context.Context, string) (*SysJobLog, error)
	// CreateJobLog 写入执行日志
	CreateJobLog(context.Context, *SysJobLog) error
	// DeleteJobLog 删除日志（批量，软删除）
	DeleteJobLog(context.Context, []string) error
	// ClearJobLog 清空日志
	ClearJobLog(context.Context) error
}

// CronjobUsecase 定时任务用例
type CronjobUsecase struct {
	repo      CronJobRepo
	scheduler *JobScheduler
	log       *log.Helper
}

func NewCronjobUsecase(repo CronJobRepo, scheduler *JobScheduler, logger log.Logger) *CronjobUsecase {
	return &CronjobUsecase{
		repo:      repo,
		scheduler: scheduler,
		log:       log.NewHelper(logger),
	}
}

func (uc *CronjobUsecase) List(ctx context.Context, req *ListJobRequest) (*ListJobResponse, error) {
	req.Normalize()
	return uc.repo.ListJob(ctx, req)
}

func (uc *CronjobUsecase) GetJob(ctx context.Context, id string) (*SysJob, error) {
	return uc.repo.GetJobByID(ctx, id)
}

func (uc *CronjobUsecase) CreateJob(ctx context.Context, req *SysJob) (*SysJob, error) {
	job := req
	job.ID = uuid.GenerateXID()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	if job.JobGroup == "" {
		job.JobGroup = "DEFAULT"
	}
	if job.JobType == "" {
		job.JobType = JobTypeHTTP
	}
	if job.MisfirePolicy == "" {
		job.MisfirePolicy = MisfirePolicyImmediately
	}
	if job.Status == "" {
		job.Status = JobStatusEnabled
	}

	// 校验 cron 表达式
	if err := uc.scheduler.ValidateCron(job.CronExpression); err != nil {
		return nil, errors.BadRequest("CRON_INVALID", err.Error())
	}

	if err := uc.repo.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	// 启用状态直接注册调度
	if job.Status == JobStatusEnabled {
		if err := uc.scheduler.AddJob(ctx, job); err != nil {
			return nil, err
		}
	}
	return job, nil
}

func (uc *CronjobUsecase) UpdateJob(ctx context.Context, req *SysJob) (*SysJob, error) {
	job := req
	job.UpdatedAt = time.Now()

	// 校验 cron 表达式
	if err := uc.scheduler.ValidateCron(job.CronExpression); err != nil {
		return nil, errors.BadRequest("CRON_INVALID", err.Error())
	}

	old, err := uc.repo.GetJobByID(ctx, job.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateJob(ctx, job); err != nil {
		return nil, err
	}

	// 重新注册调度：移除旧任务
	uc.scheduler.RemoveJob(old.ID)
	if job.Status == JobStatusEnabled {
		if err := uc.scheduler.AddJob(ctx, job); err != nil {
			return nil, err
		}
	}
	return job, nil
}

func (uc *CronjobUsecase) UpdateJobStatus(ctx context.Context, id, status string) error {
	if err := uc.repo.UpdateJobStatus(ctx, id, status); err != nil {
		return err
	}
	if status == JobStatusEnabled {
		job, err := uc.repo.GetJobByID(ctx, id)
		if err != nil {
			return err
		}
		return uc.scheduler.AddJob(ctx, job)
	}
	uc.scheduler.RemoveJob(id)
	return nil
}

func (uc *CronjobUsecase) DeleteJob(ctx context.Context, ids []string) error {
	if err := uc.repo.DeleteJob(ctx, ids); err != nil {
		return err
	}
	for _, id := range ids {
		uc.scheduler.RemoveJob(id)
	}
	return nil
}

// RunJob 立即执行一次（异步）
func (uc *CronjobUsecase) RunJob(ctx context.Context, id string) error {
	job, err := uc.repo.GetJobByID(ctx, id)
	if err != nil {
		return err
	}
	// 异步执行，不阻塞请求
	go uc.scheduler.RunJobOnce(job)
	return nil
}

func (uc *CronjobUsecase) ListJobLog(ctx context.Context, req *ListJobLogRequest) (*ListJobLogResponse, error) {
	req.Normalize()
	return uc.repo.ListJobLog(ctx, req)
}

func (uc *CronjobUsecase) GetJobLog(ctx context.Context, id string) (*SysJobLog, error) {
	return uc.repo.GetJobLogByID(ctx, id)
}

func (uc *CronjobUsecase) DeleteJobLog(ctx context.Context, ids []string) error {
	return uc.repo.DeleteJobLog(ctx, ids)
}

func (uc *CronjobUsecase) ClearJobLog(ctx context.Context) error {
	return uc.repo.ClearJobLog(ctx)
}
