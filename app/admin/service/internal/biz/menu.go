package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

type Menu struct {
	ID            string    `json:"id"`
	ParentID      string    `json:"parentID"`
	Name          string    `json:"name"`
	Code          string    `json:"code"`
	Title         string    `json:"title"`         // 标题（国际化 key）
	Type          string    `json:"type"`          // 类型: catalog / menu / iframe / link / action
	Path          string    `json:"path"`
	Icon          string    `json:"icon"`
	LinkUrl       string    `json:"linkUrl"`       // 链接地址（iframe/外链）
	Component     string    `json:"component"`
	Weight        int       `json:"weight"`
	Status        string    `json:"status"`
	AuthCode      string    `json:"authCode"`      // 权限标识
	BadgeType     string    `json:"badgeType"`     // 徽标类型: dot / text
	Badge         string    `json:"badge"`         // 徽标内容
	BadgeVariants string    `json:"badgeVariants"` // 徽标样式
	Remark        string    `json:"remark"`
	APIPermissions []*ApiPermission `json:"apiPermissions"` // 接口权限列表（仅按钮类型 action 使用）
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	Children []*Menu `json:"children"`
}

// ApiPermission 接口权限领域模型（菜单按钮绑定）
type ApiPermission struct {
	ID      string `json:"id"`      // 数据库主键（已有记录回显用；新建时为空）
	Service string `json:"service"` // 服务全限定名
	Method  string `json:"method"`  // HTTP method
	Path    string `json:"path"`    // 接口路径
	Summary string `json:"summary"` // 接口描述
}

// Code 业务唯一键：service + "|" + method + "|" + path
func (a *ApiPermission) Code() string {
	return a.Service + "|" + a.Method + "|" + a.Path
}

type MenuRepo interface {
	CreateMenu(ctx context.Context, p *Menu) error
	UpdateMenu(ctx context.Context, p *Menu) error
	UpdateMenuStatus(ctx context.Context, id, status string) error
	DeleteMenu(ctx context.Context, ids []string) error
	ListMenu(ctx context.Context, req *ListMenuRequest) ([]*Menu, error)
	// GetMenusByIDs 根据ID列表查询菜单（用于用户菜单查询）
	GetMenusByIDs(ctx context.Context, ids []string) ([]*Menu, error)
}

// ApiPermissionRepo 接口权限仓储接口
type ApiPermissionRepo interface {
	// UpsertByCodes 按业务唯一键创建或查找 ApiPermission，返回带 ID 的 biz 实体供后续关系绑定
	UpsertByCodes(ctx context.Context, items []*ApiPermission) ([]*ApiPermission, error)
	ListByCodes(ctx context.Context, codes []string) ([]*ApiPermission, error)
	ListAll(ctx context.Context) ([]*ApiPermission, error)
}

type ListMenuRequest struct {
	Name   string
	Code   string
	Status string
}

type MenuUsecase struct {
	repo MenuRepo
	log  *log.Helper
}

func NewMenuUsecase(repo MenuRepo, logger log.Logger) *MenuUsecase {
	return &MenuUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// 列表权限的tree
func (uc *MenuUsecase) MenuTree(ctx context.Context, req *ListMenuRequest) ([]*Menu, error) {
	list, err := uc.repo.ListMenu(ctx, req)
	if err != nil {
		return nil, err
	}
	tree := buildMenuTree(list)

	return tree, nil
}

// 创建权限
func (uc *MenuUsecase) CreateMenu(ctx context.Context, req *Menu) (*Menu, error) {
	permission := req

	now := time.Now()
	permission.ID = uuid.GenerateXID()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	err := uc.repo.CreateMenu(ctx, req)
	return permission, err
}

// 更新权限
func (uc *MenuUsecase) UpdateMenu(ctx context.Context, req *Menu) (*Menu, error) {
	permission := req

	now := time.Now()
	permission.UpdatedAt = now

	err := uc.repo.UpdateMenu(ctx, permission)

	return permission, err
}

// 删除权限
func (uc *MenuUsecase) DeleteMenu(ctx context.Context, ids []string) error {
	return uc.repo.DeleteMenu(ctx, ids)
}

// 更新权限状态
func (uc *MenuUsecase) UpdateMenuStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateMenuStatus(ctx, id, status)
}
