package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	securityV1 "github.com/antsurge/weaver-admin/api/gen/go/security/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SecurityService 安全中心服务（安全策略 / 在线用户 / 操作日志 / 登录日志）
type SecurityService struct {
	adminV1.UnimplementedSecurityServer

	securityUc *biz.SecurityUsecase
	onlineUc   *biz.OnlineUsecase
	logUc      *biz.LogUsecase
	log        *log.Helper
}

func NewSecurityService(securityUc *biz.SecurityUsecase, onlineUc *biz.OnlineUsecase, logUc *biz.LogUsecase, logger log.Logger) *SecurityService {
	return &SecurityService{
		securityUc: securityUc,
		onlineUc:   onlineUc,
		logUc:      logUc,
		log:        log.NewHelper(logger),
	}
}

// GetSecurityPolicy 获取安全策略
func (s *SecurityService) GetSecurityPolicy(ctx context.Context, _ *securityV1.GetSecurityPolicyRequest) (*securityV1.SecurityPolicy, error) {
	p, err := s.securityUc.GetPolicy(ctx)
	if err != nil {
		return nil, err
	}
	return toProtoSecurityPolicy(p), nil
}

// UpdateSecurityPolicy 更新安全策略
func (s *SecurityService) UpdateSecurityPolicy(ctx context.Context, req *securityV1.UpdateSecurityPolicyRequest) (*securityV1.SecurityPolicy, error) {
	p := &biz.SecurityPolicy{
		LoginFailEnabled:       req.GetLoginFailEnabled(),
		LoginFailMax:           int(req.GetLoginFailMax()),
		LoginFailWindowMinutes: int(req.GetLoginFailWindowMinutes()),
		LoginLockMinutes:       int(req.GetLoginLockMinutes()),
		IPMode:                 req.GetIpMode(),
		IPList:                 req.GetIpList(),
		AccessTokenTTLSeconds:  req.GetAccessTokenTtlSeconds(),
		RefreshTokenTTLSeconds: req.GetRefreshTokenTtlSeconds(),
		Remark:                 req.GetRemark(),
	}

	updated, err := s.securityUc.UpdatePolicy(ctx, p)
	if err != nil {
		return nil, err
	}
	return toProtoSecurityPolicy(updated), nil
}

func toProtoSecurityPolicy(p *biz.SecurityPolicy) *securityV1.SecurityPolicy {
	if p == nil {
		return nil
	}
	out := &securityV1.SecurityPolicy{
		LoginFailEnabled:       p.LoginFailEnabled,
		LoginFailMax:           int32(p.LoginFailMax),
		LoginFailWindowMinutes: int32(p.LoginFailWindowMinutes),
		LoginLockMinutes:       int32(p.LoginLockMinutes),
		IpMode:                 p.IPMode,
		IpList:                 p.IPList,
		AccessTokenTtlSeconds:  p.AccessTokenTTLSeconds,
		RefreshTokenTtlSeconds: p.RefreshTokenTTLSeconds,
		Remark:                 p.Remark,
	}
	if !p.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(p.UpdatedAt)
	}
	return out
}
