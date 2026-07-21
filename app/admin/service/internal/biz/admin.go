package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// Admin 是 biz 层的领域模型 (Domain Object)
// 这里的字段对应你 Proto 中定义的 Admin 消息
type Admin struct {
	ID        string
	RealName  string
	Username  string
	Email     string
	Phone     string
	Avatar    string
	Password  string // 用于创建时传递，查询时通常为空
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListAdminRequest struct {
	enthelper.PaginationParams
}

type ListAdminResponse struct {
	Items []*Admin
	Total int
}

type ListAdminOption struct {
	enthelper.QueryOption
}

// AdminRepo 是数据持久化层的接口定义
// 它不关心底层是 PostgreSQL 还是 Ent，只关心业务行为
type AdminRepo interface {
	ListAdmin(ctx context.Context, params *ListAdminRequest, opts ...*ListAdminOption) (*ListAdminResponse, error)
	CreateAdmin(ctx context.Context, admin *Admin) error
	UpdateAdmin(ctx context.Context, admin *Admin) error
	FindByUsername(ctx context.Context, username string) (*Admin, error)
	FindByID(ctx context.Context, id string) (*Admin, error)
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

func (uc *AdminUseCase) ListAdmin(ctx context.Context, params *ListAdminRequest) (*ListAdminResponse, error) {
	return uc.repo.ListAdmin(ctx, params)
}

func (uc *AdminUseCase) CreateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	now := time.Now()
	admin.ID = uuid.GenerateXID()
	admin.CreatedAt = now
	admin.UpdatedAt = now

	err := uc.repo.CreateAdmin(ctx, admin)
	return admin, err
}

func (uc *AdminUseCase) UpdateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	now := time.Now()
	admin.UpdatedAt = now
	err := uc.repo.UpdateAdmin(ctx, admin)

	return admin, err
}
