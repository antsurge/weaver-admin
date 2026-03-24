package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

// 岗位
type Position struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Weight    int       `json:"weight"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PositionRepo interface {
	ListPosition(context.Context, *ListPositionRequest, ...*ListPositionOption) (*ListPositionResponse, error)
	GetPosition(ctx context.Context, id string) (*Position, error)
	CreatePosition(ctx context.Context, o *Position) error
	UpdatePosition(ctx context.Context, o *Position) error
	DeletePosition(ctx context.Context, ids []string) error
	UpdatePositionStatus(ctx context.Context, id, status string) error
}

type ListPositionRequest struct {
	enthelper.PaginationParams
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type ListPositionResponse struct {
	Items []*Position
	Total int
}

type ListPositionOption struct {
	enthelper.QueryOption
}

type PositionUsecase struct {
	repo PositionRepo
	log  *log.Helper
}

func NewPositionUsecase(repo PositionRepo, logger log.Logger) *PositionUsecase {
	return &PositionUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *PositionUsecase) ListPosition(ctx context.Context, req *ListPositionRequest, opts ...*ListPositionOption) (*ListPositionResponse, error) {
	return uc.repo.ListPosition(ctx, req, opts...)
}

func (uc *PositionUsecase) GetPosition(ctx context.Context, id string) (*Position, error) {
	return uc.repo.GetPosition(ctx, id)
}

func (uc *PositionUsecase) CreatePosition(ctx context.Context, req *Position) (*Position, error) {
	position := req
	position.ID = uuid.GenerateXID()
	position.CreatedAt = time.Now()
	position.UpdatedAt = time.Now()

	err := uc.repo.CreatePosition(ctx, position)
	return position, err
}

func (uc *PositionUsecase) UpdatePosition(ctx context.Context, req *Position) (*Position, error) {
	position := req
	position.UpdatedAt = time.Now()
	err := uc.repo.UpdatePosition(ctx, position)

	return position, err
}

func (uc *PositionUsecase) DeletePosition(ctx context.Context, ids []string) error {
	return uc.repo.DeletePosition(ctx, ids)
}

func (uc *PositionUsecase) UpdatePositionStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdatePositionStatus(ctx, id, status)
}
