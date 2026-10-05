package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	opsV1 "github.com/antsurge/weaver-admin/api/gen/go/ops/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OpsService struct {
	adminV1.UnimplementedOpsServer

	cronjobUc *biz.CronjobUsecase
	monitorUc *biz.MonitorUsecase
}

func NewOpsService(cronjobUc *biz.CronjobUsecase, monitorUc *biz.MonitorUsecase) *OpsService {
	return &OpsService{
		cronjobUc: cronjobUc,
		monitorUc: monitorUc,
	}
}

// ====== 定时任务 ======

func (s *OpsService) ListJob(ctx context.Context, req *opsV1.ListJobRequest) (*opsV1.ListJobResponse, error) {
	input := biz.ListJobRequest{}
	if err := copier.Copy(&input, req); err != nil {
		return nil, err
	}

	res, err := s.cronjobUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &opsV1.ListJobResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		output.Items = make([]*opsV1.Job, 0, len(res.Data))
		for _, v := range res.Data {
			item := &opsV1.Job{}
			if err := copierx.Copy(item, v); err != nil {
				return nil, err
			}
			output.Items = append(output.Items, item)
		}
	}
	return output, nil
}

func (s *OpsService) GetJob(ctx context.Context, req *opsV1.GetJobRequest) (*opsV1.Job, error) {
	job, err := s.cronjobUc.GetJob(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}

	output := &opsV1.Job{}
	if err := copierx.Copy(output, job); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *OpsService) CreateJob(ctx context.Context, req *opsV1.CreateJobRequest) (*opsV1.Job, error) {
	input := biz.SysJob{}
	if err := copier.Copy(&input, req); err != nil {
		return nil, err
	}

	job, err := s.cronjobUc.CreateJob(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &opsV1.Job{}
	if err := copierx.Copy(output, job); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *OpsService) UpdateJob(ctx context.Context, req *opsV1.UpdateJobRequest) (*opsV1.Job, error) {
	input := biz.SysJob{}
	if err := copier.Copy(&input, req); err != nil {
		return nil, err
	}

	job, err := s.cronjobUc.UpdateJob(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &opsV1.Job{}
	if err := copierx.Copy(output, job); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *OpsService) UpdateJobStatus(ctx context.Context, req *opsV1.UpdateJobStatusRequest) (*emptypb.Empty, error) {
	err := s.cronjobUc.UpdateJobStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *OpsService) DeleteJob(ctx context.Context, req *opsV1.DeleteJobRequest) (*emptypb.Empty, error) {
	err := s.cronjobUc.DeleteJob(ctx, req.Ids)
	return nil, err
}

func (s *OpsService) RunJob(ctx context.Context, req *opsV1.RunJobRequest) (*emptypb.Empty, error) {
	err := s.cronjobUc.RunJob(ctx, req.Id)
	return nil, err
}

func (s *OpsService) GetJobLog(ctx context.Context, req *opsV1.GetJobLogRequest) (*opsV1.JobLog, error) {
	logEntry, err := s.cronjobUc.GetJobLog(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if logEntry == nil {
		return nil, nil
	}

	output := &opsV1.JobLog{}
	if err := copierx.Copy(output, logEntry); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *OpsService) ListJobLog(ctx context.Context, req *opsV1.ListJobLogRequest) (*opsV1.ListJobLogResponse, error) {
	input := biz.ListJobLogRequest{}
	if err := copier.Copy(&input, req); err != nil {
		return nil, err
	}

	res, err := s.cronjobUc.ListJobLog(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &opsV1.ListJobLogResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		output.Items = make([]*opsV1.JobLog, 0, len(res.Data))
		for _, v := range res.Data {
			item := &opsV1.JobLog{}
			if err := copierx.Copy(item, v); err != nil {
				return nil, err
			}
			output.Items = append(output.Items, item)
		}
	}
	return output, nil
}

func (s *OpsService) DeleteJobLog(ctx context.Context, req *opsV1.DeleteJobLogRequest) (*emptypb.Empty, error) {
	err := s.cronjobUc.DeleteJobLog(ctx, req.Ids)
	return nil, err
}

func (s *OpsService) ClearJobLog(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.cronjobUc.ClearJobLog(ctx)
	return nil, err
}

// ====== 系统监控 ======

// GetServerInfo 服务监控
func (s *OpsService) GetServerInfo(ctx context.Context, _ *opsV1.GetServerInfoRequest) (*opsV1.ServerInfo, error) {
	info, err := s.monitorUc.ServerInfo(ctx)
	if err != nil {
		return nil, err
	}
	out := &opsV1.ServerInfo{
		Hostname:       info.Hostname,
		Os:             info.OS,
		Arch:           info.Arch,
		GoVersion:      info.GoVersion,
		NumCPU:         int32(info.NumCPU),
		NumGoroutine:   int32(info.NumGoroutine),
		GoMaxProcs:     int32(info.GoMaxProcs),
		HeapAlloc:      info.HeapAlloc,
		HeapSys:        info.HeapSys,
		HeapInuse:      info.HeapInuse,
		StackInuse:     info.StackInuse,
		Sys:            info.Sys,
		TotalAlloc:     info.TotalAlloc,
		NumGC:          info.NumGC,
		GcPauseTotalMs: info.GCPauseTotalMs,
		UptimeSeconds:  info.UptimeSeconds,
	}
	if !info.StartTime.IsZero() {
		out.StartTime = info.StartTime.Format("2006-01-02 15:04:05")
	}
	return out, nil
}

// GetCacheInfo 缓存监控
func (s *OpsService) GetCacheInfo(ctx context.Context, _ *opsV1.GetCacheInfoRequest) (*opsV1.CacheInfo, error) {
	info, err := s.monitorUc.CacheInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &opsV1.CacheInfo{
		Version:          info.Version,
		Mode:             info.Mode,
		Role:             info.Role,
		UptimeSeconds:    info.UptimeSeconds,
		ConnectedClients: int32(info.ConnectedClients),
		UsedMemory:       info.UsedMemory,
		UsedMemoryHuman:  info.UsedMemoryHuman,
		MaxMemory:        info.MaxMemory,
		TotalKeys:        info.TotalKeys,
		KeyspaceHits:     info.KeyspaceHits,
		KeyspaceMisses:   info.KeyspaceMisses,
		HitRate:          info.HitRate,
		TotalCommands:    info.TotalCommands,
		ExpiredKeys:      info.ExpiredKeys,
		EvictedKeys:      info.EvictedKeys,
		OpsPerSec:        info.OpsPerSec,
	}, nil
}

// ListCacheKey 缓存列表
func (s *OpsService) ListCacheKey(ctx context.Context, req *opsV1.ListCacheKeyRequest) (*opsV1.ListCacheKeyResponse, error) {
	items, total, err := s.monitorUc.CacheKeys(ctx, req.GetPattern(), req.GetCurrentPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := &opsV1.ListCacheKeyResponse{Total: total, Items: make([]*opsV1.CacheKey, 0, len(items))}
	for _, v := range items {
		out.Items = append(out.Items, &opsV1.CacheKey{
			Key:        v.Key,
			RawKey:     v.RawKey,
			Type:       v.Type,
			TtlSeconds: v.TTLSeconds,
			Size:       v.Size,
			Value:      v.Value,
		})
	}
	return out, nil
}

// GetDatabaseInfo 数据库监控
func (s *OpsService) GetDatabaseInfo(ctx context.Context, _ *opsV1.GetDatabaseInfoRequest) (*opsV1.DatabaseInfo, error) {
	info, err := s.monitorUc.DatabaseInfo(ctx)
	if err != nil {
		return nil, err
	}
	out := &opsV1.DatabaseInfo{
		Driver:             info.Driver,
		Version:            info.Version,
		Database:           info.Database,
		TableCount:         int32(info.TableCount),
		MaxOpenConnections: int32(info.MaxOpenConnections),
		OpenConnections:    int32(info.OpenConnections),
		InUse:              int32(info.InUse),
		Idle:               int32(info.Idle),
		WaitCount:          info.WaitCount,
		WaitDurationMs:     info.WaitDurationMs,
		MaxIdleClosed:      info.MaxIdleClosed,
		MaxLifetimeClosed:  info.MaxLifetimeClosed,
		Tables:             make([]*opsV1.DbTableStat, 0, len(info.Tables)),
	}
	for _, t := range info.Tables {
		out.Tables = append(out.Tables, &opsV1.DbTableStat{
			Name:      t.Name,
			Rows:      t.Rows,
			DataSize:  t.DataSize,
			IndexSize: t.IndexSize,
		})
	}
	return out, nil
}

// GetApiMetrics 接口监控
func (s *OpsService) GetApiMetrics(ctx context.Context, req *opsV1.GetApiMetricsRequest) (*opsV1.GetApiMetricsResponse, error) {
	items, totalRequests, totalErrors, uptime := s.monitorUc.ApiMetrics(ctx, req.GetMethod(), req.GetPath())
	out := &opsV1.GetApiMetricsResponse{
		TotalRequests: totalRequests,
		TotalErrors:   totalErrors,
		UptimeSeconds: uptime,
		Items:         make([]*opsV1.ApiMetric, 0, len(items)),
	}
	for _, m := range items {
		out.Items = append(out.Items, &opsV1.ApiMetric{
			Method:    m.Method,
			Path:      m.Path,
			Total:     m.Total,
			Success:   m.Success,
			Error:     m.Error,
			ErrorRate: m.ErrorRate,
			AvgMs:     m.AvgMs,
			MaxMs:     m.MaxMs,
			LastMs:    m.LastMs,
		})
	}
	return out, nil
}

// GetJobMonitor 定时任务监控
func (s *OpsService) GetJobMonitor(ctx context.Context, req *opsV1.GetJobMonitorRequest) (*opsV1.JobMonitorInfo, error) {
	info, err := s.monitorUc.JobMonitor(ctx, req.GetName(), req.GetGroup())
	if err != nil {
		return nil, err
	}
	out := &opsV1.JobMonitorInfo{
		TotalJobs:    info.TotalJobs,
		EnabledJobs:  info.EnabledJobs,
		DisabledJobs: info.DisabledJobs,
		TotalRuns:    info.TotalRuns,
		SuccessRuns:  info.SuccessRuns,
		FailedRuns:   info.FailedRuns,
		SuccessRate:  info.SuccessRate,
		RunningJobs:  int64(info.RunningJobs),
		RecentRuns:   make([]*opsV1.JobRunLog, 0, len(info.RecentRuns)),
	}
	for _, v := range info.RecentRuns {
		out.RecentRuns = append(out.RecentRuns, &opsV1.JobRunLog{
			Id:        v.ID,
			JobName:   v.JobName,
			JobGroup:  v.JobGroup,
			Status:    v.Status,
			Message:   v.Message,
			StartTime: v.StartTime,
			Duration:  v.Duration,
		})
	}
	return out, nil
}
