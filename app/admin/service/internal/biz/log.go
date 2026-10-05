package biz

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// ───────────────────────── 常量区 ─────────────────────────

// 日志结果
const (
	LogStatusSuccess = "success"
	LogStatusFail    = "fail"
)

// Recorder 默认值：缓冲/批量/刷盘间隔
const (
	logBufferSize  = 4096
	logBatchSize   = 100
	logFlushEvery  = time.Second
	logWriteTimout = 5 * time.Second
)

// ───────────────────────── 领域模型 ─────────────────────────

// OperationLog 操作日志
type OperationLog struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId,omitempty"`
	AdminID   string    `json:"adminId"`
	AdminName string    `json:"adminName,omitempty"`
	Module    string    `json:"module"`
	Operation string    `json:"operation"`
	Method    string    `json:"method,omitempty"`
	Path      string    `json:"path,omitempty"`
	Summary   string    `json:"summary,omitempty"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
	Params    string    `json:"params,omitempty"`
	Status    string    `json:"status"`
	Code      string    `json:"code,omitempty"`
	Message   string    `json:"message,omitempty"`
	CostMS    int64     `json:"costMs"`
	CreatedAt time.Time `json:"createdAt"`
}

// LoginLog 登录日志
type LoginLog struct {
	ID        string    `json:"id"`
	AdminID   string    `json:"adminId,omitempty"`
	Username  string    `json:"username"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// ───────────────────────── 请求/响应 ─────────────────────────

type ListOperationLogRequest struct {
	enthelper.PaginationParams
	AdminID   string     `form:"adminId" query:"adminId"`
	AdminName string     `form:"adminName" query:"adminName"`
	Module    string     `form:"module" query:"module"`
	Operation string     `form:"operation" query:"operation"`
	Status    string     `form:"status" query:"status"`
	Keyword   string     `form:"keyword" query:"keyword"`
	StartTime *time.Time `form:"startTime" query:"startTime"`
	EndTime   *time.Time `form:"endTime" query:"endTime"`
}

type ListOperationLogResponse struct {
	Data  []*OperationLog
	Total int
}

type ListLoginLogRequest struct {
	enthelper.PaginationParams
	Username  string     `form:"username" query:"username"`
	Status    string     `form:"status" query:"status"`
	IP        string     `form:"ip" query:"ip"`
	StartTime *time.Time `form:"startTime" query:"startTime"`
	EndTime   *time.Time `form:"endTime" query:"endTime"`
}

type ListLoginLogResponse struct {
	Data  []*LoginLog
	Total int
}

// ───────────────────────── 仓储接口 ─────────────────────────

type LogRepo interface {
	// 操作日志
	CreateOperationLog(ctx context.Context, item *OperationLog) error
	CreateOperationLogBulk(ctx context.Context, items []*OperationLog) error
	ListOperationLog(ctx context.Context, req *ListOperationLogRequest) (*ListOperationLogResponse, error)
	GetOperationLog(ctx context.Context, id string) (*OperationLog, error)
	DeleteOperationLog(ctx context.Context, ids []string) error
	ClearOperationLog(ctx context.Context) error

	// 登录日志
	CreateLoginLog(ctx context.Context, item *LoginLog) error
	ListLoginLog(ctx context.Context, req *ListLoginLogRequest) (*ListLoginLogResponse, error)
	DeleteLoginLog(ctx context.Context, ids []string) error
	ClearLoginLog(ctx context.Context) error
}

// ───────────────────────── 用例 ─────────────────────────

type LogUsecase struct {
	repo      LogRepo
	adminRepo AdminRepo
	log       *log.Helper
}

func NewLogUsecase(repo LogRepo, adminRepo AdminRepo, logger log.Logger) *LogUsecase {
	return &LogUsecase{
		repo:      repo,
		adminRepo: adminRepo,
		log:       log.NewHelper(logger),
	}
}

func (uc *LogUsecase) ListOperationLog(ctx context.Context, req *ListOperationLogRequest) (*ListOperationLogResponse, error) {
	res, err := uc.repo.ListOperationLog(ctx, req)
	if err != nil {
		return nil, err
	}
	uc.fillAdminNames(ctx, res.Data...)
	return res, nil
}

func (uc *LogUsecase) GetOperationLog(ctx context.Context, id string) (*OperationLog, error) {
	item, err := uc.repo.GetOperationLog(ctx, id)
	if err != nil {
		return nil, err
	}
	if item != nil {
		uc.fillAdminNames(ctx, item)
	}
	return item, nil
}

func (uc *LogUsecase) DeleteOperationLog(ctx context.Context, ids []string) error {
	return uc.repo.DeleteOperationLog(ctx, ids)
}

func (uc *LogUsecase) ClearOperationLog(ctx context.Context) error {
	return uc.repo.ClearOperationLog(ctx)
}

func (uc *LogUsecase) ListLoginLog(ctx context.Context, req *ListLoginLogRequest) (*ListLoginLogResponse, error) {
	return uc.repo.ListLoginLog(ctx, req)
}

func (uc *LogUsecase) DeleteLoginLog(ctx context.Context, ids []string) error {
	return uc.repo.DeleteLoginLog(ctx, ids)
}

func (uc *LogUsecase) ClearLoginLog(ctx context.Context) error {
	return uc.repo.ClearLoginLog(ctx)
}

// RecordLogin 记录一次登录尝试（成功/失败都记）。
// 失败也要记，否则无法排查撞库；写失败只记日志不影响登录主流程。
func (uc *LogUsecase) RecordLogin(ctx context.Context, item *LoginLog) {
	if item == nil {
		return
	}
	if item.ID == "" {
		item.ID = uuid.GenerateXID()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if err := uc.repo.CreateLoginLog(ctx, item); err != nil {
		uc.log.Errorf("write login log failed: %v", err)
	}
}

// clientInfoFromCtx 从请求上下文提取客户端 IP 与 UA（登录日志等匿名场景使用）。
func clientInfoFromCtx(ctx context.Context) (ip, ua string) {
	// 优先从 HTTP 请求对象取；取不到再退回到 transport header
	if r := httpRequestFromCtx(ctx); r != nil {
		if v := r.Header.Get("X-Forwarded-For"); v != "" {
			if idx := strings.Index(v, ","); idx > 0 {
				ip = strings.TrimSpace(v[:idx])
			} else {
				ip = strings.TrimSpace(v)
			}
		} else if v := r.Header.Get("X-Real-IP"); v != "" {
			ip = strings.TrimSpace(v)
		} else if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
		ua = r.Header.Get("User-Agent")
	}
	if ip == "" || ua == "" {
		if tr, ok := transport.FromServerContext(ctx); ok {
			if ip == "" {
				ip = firstForwardedIP(tr.RequestHeader().Get("X-Forwarded-For"))
			}
			if ip == "" {
				ip = tr.RequestHeader().Get("X-Real-IP")
			}
			if ua == "" {
				ua = tr.RequestHeader().Get("User-Agent")
			}
		}
	}
	return
}

// httpRequestFromCtx 从请求上下文提取底层 *http.Request（HTTP 与 gRPC-gateway 都适用）。
func httpRequestFromCtx(ctx context.Context) *http.Request {
	if r, ok := khttp.RequestFromServerContext(ctx); ok {
		return r
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		if htr, ok := tr.(interface{ Request() *http.Request }); ok {
			return htr.Request()
		}
	}
	return nil
}

// firstForwardedIP 取 X-Forwarded-For 链路中第一个 IP。
func firstForwardedIP(v string) string {
	if v == "" {
		return ""
	}
	if idx := strings.Index(v, ","); idx > 0 {
		return strings.TrimSpace(v[:idx])
	}
	return strings.TrimSpace(v)
}

// fillAdminNames 批量补齐操作人姓名（列表页展示用）。
// 单页数据量小，一次批量查询即可，不需要额外缓存。
func (uc *LogUsecase) fillAdminNames(ctx context.Context, items ...*OperationLog) {
	if len(items) == 0 || uc.adminRepo == nil {
		return
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item.AdminID != "" {
			ids = append(ids, item.AdminID)
		}
	}
	if len(ids) == 0 {
		return
	}
	admins, err := uc.adminRepo.FindByIDs(ctx, ids)
	if err != nil {
		uc.log.Errorf("load admin names for operation log failed: %v", err)
		return
	}
	nameMap := make(map[string]string, len(admins))
	for _, a := range admins {
		nameMap[a.ID] = a.RealName
	}
	for _, item := range items {
		if name, ok := nameMap[item.AdminID]; ok {
			item.AdminName = name
		}
	}
}

// ───────────────────────── 异步批量落库 ─────────────────────────

// OperationLogRecorder 操作日志的进程内异步批量写入器。
//
// 设计取舍：日志不走 MQ。审计数据要求尽量不丢，而 Redis Pub/Sub 不持久化，
// MQ 抖动就会造成审计缺口；改为进程内缓冲 + 批量落库，既不影响请求耗时，
// 也不依赖任何外部组件。缓冲满时丢弃并计数（宁可丢日志也不能拖垮业务）。
type OperationLogRecorder struct {
	repo    LogRepo
	log     *log.Helper
	ch      chan *OperationLog
	stop    chan struct{}
	wg      sync.WaitGroup
	startMu sync.Mutex
	started bool
	dropped int64
}

func NewOperationLogRecorder(repo LogRepo, logger log.Logger) *OperationLogRecorder {
	return &OperationLogRecorder{
		repo: repo,
		log:  log.NewHelper(logger),
		ch:   make(chan *OperationLog, logBufferSize),
		stop: make(chan struct{}),
	}
}

// Record 非阻塞投递一条操作日志。
func (r *OperationLogRecorder) Record(item *OperationLog) {
	if item == nil {
		return
	}
	if item.ID == "" {
		item.ID = uuid.GenerateXID()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.Status == "" {
		item.Status = LogStatusSuccess
	}
	select {
	case r.ch <- item:
	default:
		atomic.AddInt64(&r.dropped, 1)
	}
}

// Dropped 返回因缓冲满被丢弃的日志条数（观测用）。
func (r *OperationLogRecorder) Dropped() int64 { return atomic.LoadInt64(&r.dropped) }

// Start 启动后台消费协程（幂等）。
func (r *OperationLogRecorder) Start() {
	r.startMu.Lock()
	defer r.startMu.Unlock()
	if r.started {
		return
	}
	r.started = true
	r.wg.Add(1)
	go r.run()
}

// Stop 停止消费并尽力把缓冲中剩余的日志刷入库。
func (r *OperationLogRecorder) Stop() {
	r.startMu.Lock()
	if !r.started {
		r.startMu.Unlock()
		return
	}
	r.started = false
	r.startMu.Unlock()

	close(r.stop)
	r.wg.Wait()
}

func (r *OperationLogRecorder) run() {
	defer r.wg.Done()

	ticker := time.NewTicker(logFlushEvery)
	defer ticker.Stop()

	batch := make([]*OperationLog, 0, logBatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		// 使用独立 context：避免请求 ctx 取消导致日志写不进去
		ctx, cancel := context.WithTimeout(context.Background(), logWriteTimout)
		err := r.repo.CreateOperationLogBulk(ctx, batch)
		cancel()
		if err != nil {
			r.log.Errorf("bulk write operation log failed, count=%d err=%v", len(batch), err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case item := <-r.ch:
			batch = append(batch, item)
			if len(batch) >= logBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-r.stop:
			// 退出前把缓冲里剩余的日志尽力写完
			for {
				select {
				case item := <-r.ch:
					batch = append(batch, item)
					if len(batch) >= logBatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}
