package biz

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// SecurityPolicyID 安全策略为全库单例，固定 ID。
const SecurityPolicyID = "default"

// IP 访问控制模式
const (
	IPModeOff       = "off"
	IPModeWhitelist = "whitelist"
	IPModeBlacklist = "blacklist"
)

// Token 有效期默认值（安全策略未配置时回退）与边界值
const (
	// DefaultAccessTokenTTLSeconds access token 默认有效期：2 小时
	DefaultAccessTokenTTLSeconds int64 = 7200
	// DefaultRefreshTokenTTLSeconds refresh token 默认有效期：7 天
	DefaultRefreshTokenTTLSeconds int64 = 604800
	// MinAccessTokenTTLSeconds access token 最小有效期：1 分钟
	MinAccessTokenTTLSeconds int64 = 60
	// MinRefreshTokenTTLSeconds refresh token 最小有效期：5 分钟
	MinRefreshTokenTTLSeconds int64 = 300
	// MaxAccessTokenTTLSeconds access token 最大有效期：30 天
	MaxAccessTokenTTLSeconds int64 = 2592000
	// MaxRefreshTokenTTLSeconds refresh token 最大有效期：365 天
	MaxRefreshTokenTTLSeconds int64 = 31536000
)

// SecurityPolicy 安全策略（登录策略 + 访问控制）
type SecurityPolicy struct {
	LoginFailEnabled       bool      // 是否启用登录失败次数限制
	LoginFailMax           int       // 窗口内允许的最大失败次数
	LoginFailWindowMinutes int       // 失败次数统计窗口（分钟）
	LoginLockMinutes       int       // 达到上限后的锁定时长（分钟）
	IPMode                 string    // off / whitelist / blacklist
	IPList                 []string  // IP 列表（支持 CIDR）
	AccessTokenTTLSeconds  int64     // access token 有效期（秒），0 回退默认值
	RefreshTokenTTLSeconds int64     // refresh token 有效期（秒），0 回退默认值
	Remark                 string    // 备注
	UpdatedAt              time.Time // 更新时间
}

// DefaultSecurityPolicy 返回安全策略默认值（数据库无记录时使用）。
func DefaultSecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		LoginFailEnabled:       false,
		LoginFailMax:           5,
		LoginFailWindowMinutes: 15,
		LoginLockMinutes:       15,
		IPMode:                 IPModeOff,
		IPList:                 []string{},
		AccessTokenTTLSeconds:  DefaultAccessTokenTTLSeconds,
		RefreshTokenTTLSeconds: DefaultRefreshTokenTTLSeconds,
	}
}

// normalize 修正缺省/非法值，保证策略可用。
func (p *SecurityPolicy) normalize() {
	if p.LoginFailMax <= 0 {
		p.LoginFailMax = 5
	}
	if p.LoginFailWindowMinutes <= 0 {
		p.LoginFailWindowMinutes = 15
	}
	if p.LoginLockMinutes <= 0 {
		p.LoginLockMinutes = 15
	}
	switch p.IPMode {
	case IPModeOff, IPModeWhitelist, IPModeBlacklist:
	default:
		p.IPMode = IPModeOff
	}
	if p.IPList == nil {
		p.IPList = []string{}
	}
	// Token 有效期边界：非法（<=0 或超范围）回退默认值
	if p.AccessTokenTTLSeconds < MinAccessTokenTTLSeconds || p.AccessTokenTTLSeconds > MaxAccessTokenTTLSeconds {
		p.AccessTokenTTLSeconds = DefaultAccessTokenTTLSeconds
	}
	if p.RefreshTokenTTLSeconds < MinRefreshTokenTTLSeconds || p.RefreshTokenTTLSeconds > MaxRefreshTokenTTLSeconds {
		p.RefreshTokenTTLSeconds = DefaultRefreshTokenTTLSeconds
	}
}

// SecurityRepo 安全策略持久化 + 登录失败计数（Redis）。
type SecurityRepo interface {
	// GetPolicy 读取策略；无记录返回 (nil, nil)
	GetPolicy(ctx context.Context) (*SecurityPolicy, error)
	// SavePolicy 保存（单例 upsert）
	SavePolicy(ctx context.Context, p *SecurityPolicy) error

	// IncrLoginFail 累加登录失败次数并返回当前值，ttl 为计数窗口
	IncrLoginFail(ctx context.Context, username string, ttl time.Duration) (int64, error)
	// ResetLoginFail 清零登录失败次数
	ResetLoginFail(ctx context.Context, username string) error
	// IsLoginLocked 查询是否处于锁定期，返回剩余锁定时间
	IsLoginLocked(ctx context.Context, username string) (bool, time.Duration, error)
	// LockLogin 锁定账号 ttl 时长
	LockLogin(ctx context.Context, username string, ttl time.Duration) error
}

// SecurityUsecase 安全策略用例
type SecurityUsecase struct {
	repo SecurityRepo
	log  *log.Helper
}

func NewSecurityUsecase(repo SecurityRepo, logger log.Logger) *SecurityUsecase {
	return &SecurityUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// GetPolicy 读取策略（不存在或读取失败时回退默认值，保证登录链路可用）。
func (uc *SecurityUsecase) GetPolicy(ctx context.Context) (*SecurityPolicy, error) {
	p, err := uc.repo.GetPolicy(ctx)
	if err != nil {
		uc.log.Errorf("get security policy failed: %v", err)
		return DefaultSecurityPolicy(), nil
	}
	if p == nil {
		return DefaultSecurityPolicy(), nil
	}
	p.normalize()
	return p, nil
}

// UpdatePolicy 更新策略（校验 + 保存 + 刷新缓存）。
func (uc *SecurityUsecase) UpdatePolicy(ctx context.Context, p *SecurityPolicy) (*SecurityPolicy, error) {
	if err := validatePolicy(p); err != nil {
		return nil, err
	}
	p.normalize()
	if err := uc.repo.SavePolicy(ctx, p); err != nil {
		return nil, err
	}
	return uc.GetPolicy(ctx)
}

// LoginLocked 登录前置检查：账号是否处于锁定期。策略未启用时恒返回未锁定。
func (uc *SecurityUsecase) LoginLocked(ctx context.Context, username string) (bool, time.Duration, error) {
	p, err := uc.GetPolicy(ctx)
	if err != nil {
		return false, 0, err
	}
	if !p.LoginFailEnabled {
		return false, 0, nil
	}
	return uc.repo.IsLoginLocked(ctx, username)
}

// RecordLoginFailure 记录一次登录失败；达到上限时锁定账号。
func (uc *SecurityUsecase) RecordLoginFailure(ctx context.Context, username string) {
	p, err := uc.GetPolicy(ctx)
	if err != nil || !p.LoginFailEnabled {
		return
	}
	window := time.Duration(p.LoginFailWindowMinutes) * time.Minute
	n, err := uc.repo.IncrLoginFail(ctx, username, window)
	if err != nil {
		uc.log.Errorf("incr login fail for %s failed: %v", username, err)
		return
	}
	if int(n) >= p.LoginFailMax {
		if err := uc.repo.LockLogin(ctx, username, time.Duration(p.LoginLockMinutes)*time.Minute); err != nil {
			uc.log.Errorf("lock login for %s failed: %v", username, err)
		}
	}
}

// ResetLoginFailure 登录成功后清零失败计数。
func (uc *SecurityUsecase) ResetLoginFailure(ctx context.Context, username string) {
	if err := uc.repo.ResetLoginFail(ctx, username); err != nil {
		uc.log.Errorf("reset login fail for %s failed: %v", username, err)
	}
}

// IsIPAllowed 判断来源 IP 是否允许访问（策略 off 时恒允许）。
func (uc *SecurityUsecase) IsIPAllowed(ctx context.Context, ip string) (bool, error) {
	p, err := uc.GetPolicy(ctx)
	if err != nil {
		return true, err
	}
	switch p.IPMode {
	case IPModeWhitelist:
		return MatchIPList(p.IPList, ip), nil
	case IPModeBlacklist:
		return !MatchIPList(p.IPList, ip), nil
	default:
		return true, nil
	}
}

// TokenTTLs 返回安全策略中配置的 token 有效期（策略未配置或读取失败时回退内置默认值）。
func (uc *SecurityUsecase) TokenTTLs(ctx context.Context) (access, refresh time.Duration) {
	p, err := uc.GetPolicy(ctx)
	if err != nil {
		uc.log.Errorf("get security policy for token ttl failed: %v", err)
		return time.Duration(DefaultAccessTokenTTLSeconds) * time.Second, time.Duration(DefaultRefreshTokenTTLSeconds) * time.Second
	}
	return time.Duration(p.AccessTokenTTLSeconds) * time.Second, time.Duration(p.RefreshTokenTTLSeconds) * time.Second
}

// MatchIPList 判断 IP 是否命中列表（支持精确 IP 与 CIDR）。
func MatchIPList(list []string, ip string) bool {
	target := net.ParseIP(strings.TrimSpace(ip))
	if target == nil {
		return false
	}
	for _, item := range list {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.Contains(item, "/") {
			if _, cidr, err := net.ParseCIDR(item); err == nil && cidr.Contains(target) {
				return true
			}
			continue
		}
		if parsed := net.ParseIP(item); parsed != nil && parsed.Equal(target) {
			return true
		}
	}
	return false
}

// validatePolicy 校验策略取值。
func validatePolicy(p *SecurityPolicy) error {
	if p.LoginFailMax < 1 || p.LoginFailMax > 100 {
		return errors.BadRequest("INVALID_LOGIN_FAIL_MAX", "登录失败次数上限需在 1-100 之间")
	}
	if p.LoginFailWindowMinutes < 1 || p.LoginFailWindowMinutes > 1440 {
		return errors.BadRequest("INVALID_LOGIN_FAIL_WINDOW", "失败统计窗口需在 1-1440 分钟之间")
	}
	if p.LoginLockMinutes < 1 || p.LoginLockMinutes > 1440 {
		return errors.BadRequest("INVALID_LOGIN_LOCK_MINUTES", "锁定时长需在 1-1440 分钟之间")
	}
	switch p.IPMode {
	case IPModeOff, IPModeWhitelist, IPModeBlacklist:
	default:
		return errors.BadRequest("INVALID_IP_MODE", "IP 模式仅支持 off/whitelist/blacklist")
	}
	for _, item := range p.IPList {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.Contains(item, "/") {
			if _, _, err := net.ParseCIDR(item); err != nil {
				return errors.BadRequest("INVALID_IP_LIST", "非法的 CIDR："+item)
			}
			continue
		}
		if net.ParseIP(item) == nil {
			return errors.BadRequest("INVALID_IP_LIST", "非法的 IP："+item)
		}
	}
	if p.AccessTokenTTLSeconds < MinAccessTokenTTLSeconds || p.AccessTokenTTLSeconds > MaxAccessTokenTTLSeconds {
		return errors.BadRequest("INVALID_ACCESS_TTL", "access token 有效期需在 60-2592000 秒之间")
	}
	if p.RefreshTokenTTLSeconds < MinRefreshTokenTTLSeconds || p.RefreshTokenTTLSeconds > MaxRefreshTokenTTLSeconds {
		return errors.BadRequest("INVALID_REFRESH_TTL", "refresh token 有效期需在 300-31536000 秒之间")
	}
	return nil
}
