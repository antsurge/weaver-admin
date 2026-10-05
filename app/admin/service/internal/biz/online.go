package biz

import (
	"context"
	"sort"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// OnlineUser 在线用户（按用户聚合的在线会话）
type OnlineUser struct {
	ID           string
	Username     string
	RealName     string
	IP           string
	UserAgent    string
	LoginAt      time.Time
	ExpireAt     time.Time
	SessionCount int
}

type OnlineUsecase struct {
	log       *log.Helper
	tokenRepo TokenRepo
	adminRepo AdminRepo
}

func NewOnlineUsecase(logger log.Logger, tokenRepo TokenRepo, adminRepo AdminRepo) *OnlineUsecase {
	return &OnlineUsecase{
		log:       log.NewHelper(logger),
		tokenRepo: tokenRepo,
		adminRepo: adminRepo,
	}
}

// ListOnlineUsers 列出当前在线用户（有效 access 会话，按用户聚合）。
func (uc *OnlineUsecase) ListOnlineUsers(ctx context.Context) ([]*OnlineUser, error) {
	sessions, err := uc.tokenRepo.ListSessions(ctx)
	if err != nil {
		uc.log.Errorf("list online sessions failed: %v", err)
		return nil, err
	}

	// 按用户聚合：会话数 + 最近一次登录的 IP/UA/时间 + 最晚到期的会话过期时间
	type agg struct {
		userID       string
		ip           string
		ua           string
		loginAt      int64
		expireAt     int64
		sessionCount int
	}
	groups := make(map[string]*agg, len(sessions))
	now := time.Now().Unix()
	for _, s := range sessions {
		a, ok := groups[s.UserID]
		if !ok {
			a = &agg{userID: s.UserID}
			groups[s.UserID] = a
		}
		a.sessionCount++
		if s.LoginAt > a.loginAt {
			a.loginAt = s.LoginAt
			a.ip = s.IP
			a.ua = s.UserAgent
		}
		if exp := now + int64(s.RemainingTTL.Seconds()); exp > a.expireAt {
			a.expireAt = exp
		}
	}

	if len(groups) == 0 {
		return nil, nil
	}

	// 批量补全用户信息
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	admins, err := uc.adminRepo.FindByIDs(ctx, ids)
	if err != nil {
		uc.log.Errorf("batch query admins failed: %v", err)
		return nil, err
	}
	byID := make(map[string]*Admin, len(admins))
	for _, a := range admins {
		byID[a.ID] = a
	}

	out := make([]*OnlineUser, 0, len(groups))
	for _, a := range groups {
		u := &OnlineUser{
			ID:           a.userID,
			IP:           a.ip,
			UserAgent:    a.ua,
			LoginAt:      time.Unix(a.loginAt, 0),
			ExpireAt:     time.Unix(a.expireAt, 0),
			SessionCount: a.sessionCount,
		}
		if admin, ok := byID[a.userID]; ok {
			u.Username = admin.Username
			u.RealName = admin.RealName
		}
		out = append(out, u)
	}

	// 最近登录的排前面
	sort.Slice(out, func(i, j int) bool { return out[i].LoginAt.After(out[j].LoginAt) })
	return out, nil
}

// ForceLogout 强制下线：踢掉指定用户的全部在线会话（access + refresh）。
func (uc *OnlineUsecase) ForceLogout(ctx context.Context, userID string) error {
	if err := uc.tokenRepo.DeleteByUser(ctx, userID); err != nil {
		uc.log.Errorf("force logout user %s failed: %v", userID, err)
		return err
	}
	return nil
}
