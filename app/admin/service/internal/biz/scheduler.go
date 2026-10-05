package biz

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/robfig/cron/v3"
)

// JobLocker 分布式锁抽象（由 data 层基于 redis 实现，biz 不依赖具体客户端）
type JobLocker interface {
	// TryLock 尝试获取锁。成功返回 true 与解锁函数；失败返回 false。
	// key 为锁标识，ttl 为锁自动过期时间（防止持有者异常退出导致死锁）。
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, func(), error)
}

// JobExecutor 任务执行器：负责真正执行任务并返回执行结果。
// 默认实现为 HTTP 调用（HTTPExecutor）；未来可扩展内部方法调用等。
type JobExecutor interface {
	// Execute 执行任务，返回执行信息（成功为响应体，失败为错误信息）
	Execute(ctx context.Context, job *SysJob) (string, error)
}

// HTTPExecutor 基于 HTTP 回调的任务执行器
type HTTPExecutor struct {
	client *http.Client
}

func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *HTTPExecutor) Execute(ctx context.Context, job *SysJob) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, job.InvokeTarget, nil)
	if err != nil {
		return "", fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Job-Id", job.ID)
	req.Header.Set("X-Job-Name", job.Name)

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP 调用失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP 调用失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应体（限制大小）
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 512)
	limited := 4096
	for limited > 0 {
		n, err := resp.Body.Read(tmp)
		if n > 0 {
			appendLen := n
			if appendLen > limited {
				appendLen = limited
			}
			buf = append(buf, tmp[:appendLen]...)
			limited -= appendLen
		}
		if err != nil {
			break
		}
	}
	return strings.TrimSpace(string(buf)), nil
}

// JobScheduler 定时任务调度器
type JobScheduler struct {
	cron     *cron.Cron
	executor JobExecutor
	locker   JobLocker
	repo     CronJobRepo

	mu     sync.RWMutex
	entrys map[string]cron.EntryID // jobID -> entryID

	lockTTL  time.Duration
	logDebug func(format string, args ...interface{})
}

// NewJobScheduler 创建调度器（默认 HTTP 执行器；locker 可为空表示不启用分布式锁）
func NewJobScheduler(repo CronJobRepo, locker JobLocker, executor JobExecutor) *JobScheduler {
	if executor == nil {
		executor = NewHTTPExecutor()
	}
	s := &JobScheduler{
		cron:     cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger))),
		executor: executor,
		locker:   locker,
		repo:     repo,
		entrys:   make(map[string]cron.EntryID),
		lockTTL:  2 * time.Minute,
	}
	return s
}

// Start 加载所有启用任务并启动调度（由 Kratos server 生命周期驱动）
func (s *JobScheduler) Start(ctx context.Context) error {
	jobs, err := s.repo.ListEnabledJobs(ctx)
	if err != nil {
		return fmt.Errorf("加载启用任务失败: %w", err)
	}
	for _, job := range jobs {
		if err := s.AddJob(ctx, job); err != nil {
			s.logf("注册任务失败 job=%s err=%v", job.Name, err)
		}
	}
	s.cron.Start()
	return nil
}

// Stop 停止调度
func (s *JobScheduler) Stop(ctx context.Context) error {
	s.cron.Stop()
	return nil
}

// parserWithSeconds 支持 6 字段（含秒）的 cron 解析器，
// 与 cron.New(cron.WithSeconds()) 行为一致。
var parserWithSeconds = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// ValidateCron 校验 cron 表达式（兼容 5 字段标准式与 6 字段含秒式）
func (s *JobScheduler) ValidateCron(expr string) error {
	if _, err := cron.ParseStandard(expr); err == nil {
		return nil
	}
	if _, err := parserWithSeconds.Parse(expr); err != nil {
		return fmt.Errorf("cron 表达式非法: %v", err)
	}
	return nil
}

// AddJob 注册任务到调度器
func (s *JobScheduler) AddJob(ctx context.Context, job *SysJob) error {
	s.RemoveJob(job.ID)

	spec, err := s.normalizeSpec(job.CronExpression)
	if err != nil {
		return err
	}

	entryID, err := s.cron.AddFunc(spec, func() {
		s.RunJobOnce(job)
	})
	if err != nil {
		return fmt.Errorf("注册调度失败: %w", err)
	}

	s.mu.Lock()
	s.entrys[job.ID] = entryID
	s.mu.Unlock()

	// 计算并更新下次执行时间
	if next := s.nextRunTime(spec); next != nil {
		if err := s.repo.UpdateNextRunTime(ctx, job.ID, next); err != nil {
			s.logf("更新下次执行时间失败 job=%s err=%v", job.Name, err)
		}
	}
	return nil
}

// RemoveJob 移除任务
func (s *JobScheduler) RemoveJob(jobID string) {
	s.mu.Lock()
	entryID, ok := s.entrys[jobID]
	if ok {
		s.cron.Remove(entryID)
		delete(s.entrys, jobID)
	}
	s.mu.Unlock()
}

// RunJobOnce 立即执行一次任务
func (s *JobScheduler) RunJobOnce(job *SysJob) {
	// 并发控制：不允许并发且正在执行时直接跳过
	if !job.Concurrent {
		if s.isRunning(job.ID) {
			s.logf("任务 %s 正在执行，跳过本次触发", job.Name)
			return
		}
		s.setRunning(job.ID)
		defer s.clearRunning(job.ID)
	}

	// 分布式锁：确保多实例下只有一个实例执行
	if s.locker != nil {
		lockKey := fmt.Sprintf("weaver:cronjob:lock:%s", job.ID)
		ok, unlock, err := s.locker.TryLock(context.Background(), lockKey, s.lockTTL)
		if err != nil {
			s.writeLog(job, false, fmt.Sprintf("获取分布式锁失败: %v", err), 0)
			return
		}
		if !ok {
			s.logf("任务 %s 已被其他实例执行，跳过", job.Name)
			return
		}
		defer unlock()
	}

	start := time.Now()
	msg, err := s.executor.Execute(context.Background(), job)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		s.writeLog(job, false, err.Error(), duration)
		return
	}
	s.writeLog(job, true, msg, duration)
}

// 运行中标记（并发控制，仅本实例内有效）
var runningMu sync.Mutex
var runningJobs = map[string]bool{}

// RunningCount 返回当前正在执行的任务数（供监控页面展示）。
func (s *JobScheduler) RunningCount() int {
	runningMu.Lock()
	defer runningMu.Unlock()
	return len(runningJobs)
}

func (s *JobScheduler) isRunning(jobID string) bool {
	runningMu.Lock()
	defer runningMu.Unlock()
	return runningJobs[jobID]
}

func (s *JobScheduler) setRunning(jobID string) {
	runningMu.Lock()
	runningJobs[jobID] = true
	runningMu.Unlock()
}

func (s *JobScheduler) clearRunning(jobID string) {
	runningMu.Lock()
	delete(runningJobs, jobID)
	runningMu.Unlock()
}

// writeLog 写入执行日志
func (s *JobScheduler) writeLog(job *SysJob, success bool, message string, duration int64) {
	now := time.Now()
	logEntry := &SysJobLog{
		ID:             uuid.GenerateXID(),
		TenantID:       job.TenantID,
		JobID:          job.ID,
		JobName:        job.Name,
		JobGroup:       job.JobGroup,
		InvokeTarget:   job.InvokeTarget,
		CronExpression: job.CronExpression,
		Status:         "success",
		// 统一截断：成功时 message 为接口响应体（可能很长），超长会导致
		// job_message 字段校验失败、日志整体写不进去（无论成功失败）。
		JobMessage: truncateRunes(message, 2000),
		StartTime:  &now,
		EndTime:    &now,
		Duration:   duration,
		CreatedAt:  now,
	}
	if !success {
		logEntry.Status = "fail"
	}
	if err := s.repo.CreateJobLog(context.Background(), logEntry); err != nil {
		s.logf("写入任务日志失败 job=%s err=%v", job.Name, err)
	}
}

// truncateRunes 按字符（rune）截断，避免超出字段长度限制或截出非法 UTF-8。
func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// normalizeSpec 兼容 5/6 段 cron 表达式，统一转成 6 段（含秒）
func (s *JobScheduler) normalizeSpec(expr string) (string, error) {
	fields := strings.Fields(expr)
	switch len(fields) {
	case 5:
		// 5 段标准 cron：秒固定为 0
		return "0 " + expr, nil
	case 6:
		if _, err := parserWithSeconds.Parse(expr); err != nil {
			return "", fmt.Errorf("cron 表达式非法: %s", expr)
		}
		return expr, nil
	default:
		return "", fmt.Errorf("cron 表达式字段数量非法: %s", expr)
	}
}

// nextRunTime 计算下次执行时间
func (s *JobScheduler) nextRunTime(spec string) *time.Time {
	schedule, err := parserWithSeconds.Parse(spec)
	if err != nil {
		return nil
	}
	t := schedule.Next(time.Now())
	return &t
}

func (s *JobScheduler) logf(format string, args ...interface{}) {
	if s.logDebug != nil {
		s.logDebug(format, args...)
	}
}

// SetLogDebug 注入日志函数（由 service 层/依赖注入提供）
func (s *JobScheduler) SetLogDebug(fn func(format string, args ...interface{})) {
	s.logDebug = fn
}
