package biz

import (
	"context"
	"math/rand"
	"strings"
	"time"

	authenticationV1 "github.com/antsurge/weaver-admin/api/gen/go/authentication/service/v1"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/mojocn/base64Captcha"
)

type CaptchaRepo interface {
	// 保存验证码（带过期时间）
	Save(ctx context.Context, id, code string, ttl time.Duration) error

	// 获取验证码内容
	Get(ctx context.Context, id string) (string, error)

	// 删除验证码
	Delete(ctx context.Context, id string) error
}

type CaptchaUsecase struct {
	repo CaptchaRepo
	log  *log.Helper
}

func NewCaptchaUsecase(
	repo CaptchaRepo,
	logger log.Logger,
) *CaptchaUsecase {
	return &CaptchaUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// 获取验证码
func (u *CaptchaUsecase) GetCaptcha(ctx context.Context) (*authenticationV1.GetCaptchaResponse, error) {
	// 生成验证码
	code := strings.ToLower(generateCaptchaCode(4))
	captchaID := generateCaptchaID()

	// 生成图片 base64
	driver := base64Captcha.NewDriverString(68, 200, 0, 0, 0, "", nil, nil, nil)
	b64s, err := driver.DrawCaptcha(code)
	if err != nil {
		// 记录日志
		u.log.Errorf("generate captcha image error: %v", err)
		return nil, authenticationV1.ErrorGenerateCaptchaFail("GENERATE_CAPTCHA_FAIL")
	}

	// 保存到 Redis，过期时间为 2 分钟
	if err := u.repo.Save(ctx, captchaID, code, 2*time.Minute); err != nil {
		// 记录日志
		u.log.Errorf("save captcha error: %v", err)
		return nil, authenticationV1.ErrorGenerateCaptchaFail("GENERATE_CAPTCHA_FAIL")
	}

	// 返回 Proto 响应
	return &authenticationV1.GetCaptchaResponse{
		CaptchaId:   captchaID,
		ImageBase64: b64s.EncodeB64string(),
	}, nil
}

// 生成随机验证码
func generateCaptchaCode(length int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	// 使用局部随机源
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteByte(chars[r.Intn(len(chars))])
	}
	return b.String()
}

// 生成 captchaId
func generateCaptchaID() string {
	return uuid.NewString()
}
