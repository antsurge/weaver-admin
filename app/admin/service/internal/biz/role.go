package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type Role struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Remark    string    `json:"remark"`
	Weight    int       `json:"weight"`
	Status    string    `json:"status"`
	DataScope string    `json:"dataScope"`
	IsSystem  bool      `json:"isSystem"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	MenuIDs   []string  `json:"menuIds"`
}

type RoleRepo interface {
	CreateMenu(context.Context, *Role) error
}

type RoleUseCase struct {
	repo RoleRepo
	log  *log.Helper
}

func NewRoleUseCase(repo RoleRepo, logger log.Logger) *RoleUseCase {
	return &RoleUseCase{repo: repo, log: log.NewHelper(logger)}
}
