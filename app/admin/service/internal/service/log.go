package service

import (
	"context"
	"time"

	securityV1 "github.com/antsurge/weaver-admin/api/gen/go/security/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 日志查询方法（采集见 pkg/middleware/oplog 与登录埋点），归属安全中心服务。

// ListOperationLog 操作日志列表
func (s *SecurityService) ListOperationLog(ctx context.Context, req *securityV1.ListOperationLogRequest) (*securityV1.ListOperationLogResponse, error) {
	input := &biz.ListOperationLogRequest{
		AdminID:   req.GetAdminId(),
		AdminName: req.GetAdminName(),
		Module:    req.GetModule(),
		Operation: req.GetOperation(),
		Status:    req.GetStatus(),
		Keyword:   req.GetKeyword(),
		StartTime: parseTime(req.GetStartTime()),
		EndTime:   parseTime(req.GetEndTime()),
	}
	input.CurrentPage = int(req.GetCurrentPage())
	input.PageSize = int(req.GetPageSize())

	res, err := s.logUc.ListOperationLog(ctx, input)
	if err != nil {
		return nil, err
	}

	out := &securityV1.ListOperationLogResponse{Total: int64(res.Total)}
	out.Items = make([]*securityV1.OperationLog, 0, len(res.Data))
	for _, item := range res.Data {
		out.Items = append(out.Items, toPbOperationLog(item))
	}
	return out, nil
}

// GetOperationLog 操作日志详情
func (s *SecurityService) GetOperationLog(ctx context.Context, req *securityV1.GetOperationLogRequest) (*securityV1.OperationLog, error) {
	item, err := s.logUc.GetOperationLog(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return toPbOperationLog(item), nil
}

// DeleteOperationLog 删除操作日志
func (s *SecurityService) DeleteOperationLog(ctx context.Context, req *securityV1.DeleteOperationLogRequest) (*emptypb.Empty, error) {
	if err := s.logUc.DeleteOperationLog(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ClearOperationLog 清空操作日志
func (s *SecurityService) ClearOperationLog(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.logUc.ClearOperationLog(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListLoginLog 登录日志列表
func (s *SecurityService) ListLoginLog(ctx context.Context, req *securityV1.ListLoginLogRequest) (*securityV1.ListLoginLogResponse, error) {
	input := &biz.ListLoginLogRequest{
		Username:  req.GetUsername(),
		Status:    req.GetStatus(),
		IP:        req.GetIp(),
		StartTime: parseTime(req.GetStartTime()),
		EndTime:   parseTime(req.GetEndTime()),
	}
	input.CurrentPage = int(req.GetCurrentPage())
	input.PageSize = int(req.GetPageSize())

	res, err := s.logUc.ListLoginLog(ctx, input)
	if err != nil {
		return nil, err
	}

	out := &securityV1.ListLoginLogResponse{Total: int64(res.Total)}
	out.Items = make([]*securityV1.LoginLog, 0, len(res.Data))
	for _, item := range res.Data {
		out.Items = append(out.Items, toPbLoginLog(item))
	}
	return out, nil
}

// DeleteLoginLog 删除登录日志
func (s *SecurityService) DeleteLoginLog(ctx context.Context, req *securityV1.DeleteLoginLogRequest) (*emptypb.Empty, error) {
	if err := s.logUc.DeleteLoginLog(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ClearLoginLog 清空登录日志
func (s *SecurityService) ClearLoginLog(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.logUc.ClearLoginLog(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ───────────────────────── 转换 ─────────────────────────

func toPbOperationLog(v *biz.OperationLog) *securityV1.OperationLog {
	if v == nil {
		return nil
	}
	return &securityV1.OperationLog{
		Id:        v.ID,
		AdminId:   v.AdminID,
		AdminName: v.AdminName,
		Module:    v.Module,
		Operation: v.Operation,
		Method:    v.Method,
		Path:      v.Path,
		Summary:   v.Summary,
		Ip:        v.IP,
		UserAgent: v.UserAgent,
		Params:    v.Params,
		Status:    v.Status,
		Code:      v.Code,
		Message:   v.Message,
		CostMs:    v.CostMS,
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
	}
}

func toPbLoginLog(v *biz.LoginLog) *securityV1.LoginLog {
	if v == nil {
		return nil
	}
	return &securityV1.LoginLog{
		Id:        v.ID,
		AdminId:   v.AdminID,
		Username:  v.Username,
		Ip:        v.IP,
		UserAgent: v.UserAgent,
		Status:    v.Status,
		Reason:    v.Reason,
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
	}
}

// parseTime 解析 RFC3339 时间，空串或非法格式返回 nil（即不过滤）。
func parseTime(v string) *time.Time {
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}
