package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

type Menu struct {
	ID             string           `json:"id"`
	ParentID       string           `json:"parentID"`
	Name           string           `json:"name"`
	Code           string           `json:"code"`
	Title          string           `json:"title"` // 标题（国际化 key）
	Type           string           `json:"type"`  // 类型: catalog / menu / iframe / link / action
	Path           string           `json:"path"`
	Icon           string           `json:"icon"`
	LinkUrl        string           `json:"linkUrl"` // 链接地址（iframe/外链）
	Component      string           `json:"component"`
	Weight         int              `json:"weight"`
	Status         string           `json:"status"`
	AuthCode       string           `json:"authCode"`      // 权限标识
	BadgeType      string           `json:"badgeType"`     // 徽标类型: dot / text
	Badge          string           `json:"badge"`         // 徽标内容
	BadgeVariants  string           `json:"badgeVariants"` // 徽标样式
	Remark         string           `json:"remark"`
	APIPermissions []*ApiPermission `json:"apiPermissions"` // 接口权限列表（仅按钮类型 action 使用）
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`

	Children []*Menu `json:"children"`
}

// ApiPermission 接口权限领域模型（菜单按钮绑定）。
// 绑定关系以 api_interface.code 为稳定键存储（menu_api_permission 表），
// 接口信息实时从 api_interface 反查，因此不依赖快照 id。
type ApiPermission struct {
	Service string `json:"service"` // 服务全限定名
	Tag     string `json:"tag"`     // OpenAPI 标签（前端分组回显用）
	Method  string `json:"method"`  // HTTP method
	Path    string `json:"path"`    // 接口路径
	Summary string `json:"summary"` // 接口描述
	Code    string `json:"code"`    // 业务唯一键 service|METHOD|path
}

// CodeKey 业务唯一键：service + "|" + method + "|" + path（与 api_interface.code 一致）
func (a *ApiPermission) CodeKey() string {
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
	// GetMenuByID 根据ID查询菜单详情（含接口权限）
	GetMenuByID(ctx context.Context, id string) (*Menu, error)
}

type ListMenuRequest struct {
	Name   string
	Code   string
	Status string
	Type   string
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

// 获取权限详情（含接口权限）
func (uc *MenuUsecase) GetMenu(ctx context.Context, id string) (*Menu, error) {
	return uc.repo.GetMenuByID(ctx, id)
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

	// status 兜底：ent 枚举校验（enabled/disabled）不允许空值
	if permission.Status == "" {
		permission.Status = "enabled"
	}
	if permission.Type == "" {
		permission.Type = "menu"
	}

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

	// status 兜底：ent 枚举校验（enabled/disabled）不允许空值
	if permission.Status == "" {
		permission.Status = "enabled"
	}

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

// buildMenuTree 将菜单列表构建为树形结构。
// 防环：若菜单数据存在循环 parent 引用（如 A.parent=B 且 B.parent=A），
// 直接互相挂载会构造出无限嵌套的环状树，导致后续任何递归遍历（菜单树查询、
// 代码生成落库菜单、删除菜单树等）无限递归直至栈溢出（runtime: goroutine stack
// exceeds ...）。因此在挂载前向上追溯 parent 链，检测到会形成环的节点
// 一律提升为根节点，从源头保证构建出的树永远无环。
func buildMenuTree(perms []*Menu) []*Menu {
	nodeMap := make(map[string]*Menu)

	// 先创建所有节点
	for _, p := range perms {
		nodeMap[p.ID] = p
	}

	var roots []*Menu

	// 构建树
	for _, p := range perms {
		node := nodeMap[p.ID]

		if p.ParentID == "" {
			roots = append(roots, node)
			continue
		}

		parent, ok := nodeMap[p.ParentID]
		if ok && !wouldFormCycle(node, parent, nodeMap) {
			parent.Children = append(parent.Children, node)
		} else {
			// 找不到父节点或挂载后形成环，当作根节点
			roots = append(roots, node)
		}
	}

	return roots
}

// wouldFormCycle 判断把 child 挂载到 parent 下是否会形成环。
// 方法：从 parent 出发沿 parentID 链向上追溯，若途中回到 child 自身，则挂载后必然成环。
func wouldFormCycle(child, parent *Menu, nodeMap map[string]*Menu) bool {
	if child == nil || parent == nil {
		return false
	}
	seen := make(map[string]bool, 8)
	cur := parent
	for cur != nil {
		if cur.ID == child.ID || seen[cur.ID] {
			return true
		}
		seen[cur.ID] = true
		if cur.ParentID == "" {
			return false
		}
		next, ok := nodeMap[cur.ParentID]
		if !ok {
			return false
		}
		cur = next
	}
	return false
}
