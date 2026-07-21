package data

import (
	"context"
	"fmt"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"time"
)

type captchaRepo struct {
	data *Data
	log  *log.Helper
}

func NewCaptchaRepo(data *Data, logger log.Logger) biz.CaptchaRepo {
	return &captchaRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *captchaRepo) Save(ctx context.Context, id, code string, ttl time.Duration) error {
	key := r.key(id)
	err := r.data.redis.Set(ctx, key, code, ttl).Err()
	if err != nil {
		r.log.Errorf("captcha save error: %v", err)
	}
	return err
}

func (r *captchaRepo) Get(ctx context.Context, id string) (string, error) {
	key := r.key(id)
	val, err := r.data.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("captcha not found or expired")
	}
	if err != nil {
		r.log.Errorf("captcha get error: %v", err)
		return "", err
	}
	return val, nil
}

// Delete 从 Redis 删除验证码
func (r *captchaRepo) Delete(ctx context.Context, id string) error {
	key := r.key(id)
	err := r.data.redis.Del(ctx, key).Err()
	if err != nil {
		r.log.Errorf("captcha delete error: %v", err)
	}
	return err
}

// key 生成 Redis Key
func (r *captchaRepo) key(id string) string {
	return fmt.Sprintf("captcha:%s", id)
}
