package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type tokenRepo struct {
	data *Data
	log  *log.Helper
	conf *conf.Data
}

func NewTokenRepo(data *Data, logger log.Logger, c *conf.Data) biz.TokenRepo {
	return &tokenRepo{
		data: data,
		log:  log.NewHelper(logger),
		conf: c,
	}
}

// tokenSessionPayload Redis value 的 JSON 结构（与 biz.TokenSession 对应）
type tokenSessionPayload struct {
	UserID    string `json:"userID"`
	Type      string `json:"type"`
	IP        string `json:"ip"`
	UserAgent string `json:"ua"`
	LoginAt   int64  `json:"loginAt"`
}

func (r *tokenRepo) Save(ctx context.Context, token string, s *biz.TokenSession, ttl time.Duration) error {
	payload, err := json.Marshal(&tokenSessionPayload{
		UserID:    s.UserID,
		Type:      s.Type,
		IP:        s.IP,
		UserAgent: s.UserAgent,
		LoginAt:   s.LoginAt,
	})
	if err != nil {
		r.log.Errorf("token marshal error: %v", err)
		return err
	}
	key := r.key(token)
	err = r.data.redis.Set(ctx, key, payload, ttl).Err()
	if err != nil {
		r.log.Errorf("token save error: %v", err)
	}
	return err
}

func (r *tokenRepo) Get(ctx context.Context, token string) (string, error) {
	key := r.key(token)
	val, err := r.data.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found or expired")
	}
	if err != nil {
		r.log.Errorf("token get error: %v", err)
		return "", err
	}
	// 兼容历史纯 userID 值；新值一律为 JSON
	var p tokenSessionPayload
	if err := json.Unmarshal([]byte(val), &p); err == nil && p.UserID != "" {
		return p.UserID, nil
	}
	return val, nil
}

func (r *tokenRepo) Delete(ctx context.Context, token string) error {
	key := r.key(token)
	err := r.data.redis.Del(ctx, key).Err()
	if err != nil {
		r.log.Errorf("token delete error: %v", err)
	}
	return err
}

func (r *tokenRepo) Exists(ctx context.Context, token string) (bool, error) {
	key := r.key(token)
	count, err := r.data.redis.Exists(ctx, key).Result()
	if err != nil {
		r.log.Errorf("token exists check error: %v", err)
		return false, err
	}
	return count > 0, nil
}

// ListSessions 扫描 Redis 中全部 token key，返回 access 会话列表（在线用户）。
// refresh token 不视为在线会话。
func (r *tokenRepo) ListSessions(ctx context.Context) ([]*biz.TokenSession, error) {
	keys, err := r.scanKeys(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}

	// 批量取 value，减少网络往返
	vals, err := r.data.redis.MGet(ctx, keys...).Result()
	if err != nil {
		r.log.Errorf("token mget error: %v", err)
		return nil, err
	}

	sessions := make([]*biz.TokenSession, 0, len(keys))
	for i, key := range keys {
		if i >= len(vals) || vals[i] == nil {
			continue
		}
		val, ok := vals[i].(string)
		if !ok || val == "" {
			continue
		}
		var p tokenSessionPayload
		if err := json.Unmarshal([]byte(val), &p); err != nil || p.UserID == "" {
			continue
		}
		// 只统计 access 会话
		if p.Type != "" && p.Type != "access" {
			continue
		}
		remaining, err := r.data.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		if remaining <= 0 {
			// TTL 已过期但 key 尚未被回收，忽略
			continue
		}
		sessions = append(sessions, &biz.TokenSession{
			Token:        strings.TrimPrefix(key, r.prefix()),
			UserID:       p.UserID,
			Type:         "access",
			IP:           p.IP,
			UserAgent:    p.UserAgent,
			LoginAt:      p.LoginAt,
			RemainingTTL: remaining,
		})
	}
	return sessions, nil
}

// DeleteByUser 删除指定用户的全部 token（强制下线）。
func (r *tokenRepo) DeleteByUser(ctx context.Context, userID string) error {
	keys, err := r.scanKeys(ctx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	vals, err := r.data.redis.MGet(ctx, keys...).Result()
	if err != nil {
		r.log.Errorf("token mget error: %v", err)
		return err
	}
	toDelete := make([]string, 0, len(keys))
	for i, key := range keys {
		if i >= len(vals) || vals[i] == nil {
			continue
		}
		val, ok := vals[i].(string)
		if !ok || val == "" {
			continue
		}
		var p tokenSessionPayload
		if err := json.Unmarshal([]byte(val), &p); err != nil {
			continue
		}
		if p.UserID == userID {
			toDelete = append(toDelete, key)
		}
	}
	if len(toDelete) == 0 {
		return nil
	}
	if err := r.data.redis.Del(ctx, toDelete...).Err(); err != nil {
		r.log.Errorf("token delete by user error: %v", err)
		return err
	}
	return nil
}

func (r *tokenRepo) scanKeys(ctx context.Context) ([]string, error) {
	var keys []string
	iter := r.data.redis.Scan(ctx, 0, r.prefix()+"*", 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		r.log.Errorf("token scan error: %v", err)
		return nil, err
	}
	return keys, nil
}

func (r *tokenRepo) prefix() string {
	return fmt.Sprintf("%stoken:", r.conf.Redis.KeyPrefix)
}

func (r *tokenRepo) key(token string) string {
	return r.prefix() + token
}
