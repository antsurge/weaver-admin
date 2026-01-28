package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"time"
)

// Admin 是 biz 层的领域模型 (Domain Object)
// 这里的字段对应你 Proto 中定义的 Admin 消息
type Admin struct {
	ID         string
	Name       string
	Username   string
	Email      string
	Phone      string
	Avatar     string
	Password   string // 用于创建时传递，查询时通常为空
	CreateTime time.Time
	UpdateTime time.Time
}

// AdminRepo 是数据持久化层的接口定义
// 它不关心底层是 PostgreSQL 还是 Ent，只关心业务行为
type AdminRepo interface {
	CreateAdmin(ctx context.Context, a *Admin) (*Admin, error)
	FindByUsername(ctx context.Context, username string) (*Admin, error)
	//GetAdmin(ctx context.Context, id string) (*Admin, error)
	//UpdateAdmin(ctx context.Context, a *Admin) (*Admin, error)
	//DeleteAdmin(ctx context.Context, id string) error
}

// AdminUseCase 是业务逻辑控制器
type AdminUseCase struct {
	repo AdminRepo
	log  *log.Helper
}

// NewAdminUseCase 构造函数，由 Wire 注入 Repo
func NewAdminUseCase(repo AdminRepo, logger log.Logger) *AdminUseCase {
	return &AdminUseCase{repo: repo, log: log.NewHelper(logger)}
}
