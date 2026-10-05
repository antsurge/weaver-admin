package biz

import (
	"context"
	crand "crypto/rand"
	"math/big"
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

/**
 * GetCaptcha 获取图形验证码。
 * 流程：生成随机验证码 -> 绘制干扰图片 -> 保存至 Redis（TTL 2 分钟）-> 返回图片 base64 与验证码 ID。
 */
func (u *CaptchaUsecase) GetCaptcha(ctx context.Context) (*authenticationV1.GetCaptchaResponse, error) {
	// 生成验证码，统一转为小写，校验时不区分大小写
	code, err := generateCaptchaCode(4)
	if err != nil {
		u.log.Errorf("generate captcha code error: %v", err)
		return nil, authenticationV1.ErrorGenerateCaptchaFail("GENERATE_CAPTCHA_FAIL")
	}
	code = strings.ToLower(code)
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

/**
 * generateCaptchaCode 生成指定长度的随机验证码。
 * 字符集刻意排除了易混淆字符（I/O/0/1），兼顾可读性与安全性。
 * 随机源使用 crypto/rand（操作系统加密安全随机源）而非 math/rand：
 * math/rand 以时间戳为种子、PRNG 状态可重建，会导致验证码可被预测，
 * 从而绕过登录/注册等接口的防爆破校验。
 */
func generateCaptchaCode(length int) (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	max := big.NewInt(int64(len(chars)))
	var b strings.Builder
	b.Grow(length)
	for i := 0; i < length; i++ {
		n, err := crand.Int(crand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(chars[n.Int64()])
	}
	return b.String(), nil
}

/**
 * generateCaptchaID 生成验证码唯一标识。
 * 使用 UUIDv4（底层基于 crypto/rand），作为 Redis 中验证码的 key，无法被猜测或枚举。
 */
func generateCaptchaID() string {
	return uuid.NewString()
}
