package data

import (
	"context"
	"fmt"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"time"
)

type tokenRepo struct {
	data *Data
	log  *log.Helper
}

func NewTokenRepo(data *Data, logger log.Logger) biz.TokenRepo {
	return &tokenRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *tokenRepo) Save(ctx context.Context, token string, userID string, ttl time.Duration) error {
	key := r.key(token)
	err := r.data.redis.Set(ctx, key, userID, ttl).Err()
	if err != nil {
		r.log.Errorf("token save error: %v", err)
	}
	return err
}

func (r *tokenRepo) Get(ctx context.Context, token string) (string, error) {
	key := r.key(token)
	userID, err := r.data.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found or expired")
	}
	if err != nil {
		r.log.Errorf("token get error: %v", err)
		return "", err
	}
	return userID, nil
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

func (r *tokenRepo) key(token string) string {
	return fmt.Sprintf("token:%s", token)
}
