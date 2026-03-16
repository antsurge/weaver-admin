package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type Position struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Weight      int       `json:"weight"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PositionRepo interface {
	ListPosition(ctx context.Context, req *ListPositionRequest) (*ListPositionResponse, error)
	CreatePosition(ctx context.Context, o *Position) error
	UpdatePosition(ctx context.Context, o *Position) error
	DeletePosition(ctx context.Context, ids []string) error
	UpdatePositionStatus(ctx context.Context, id, status string) error
}

type ListPositionRequest struct {
	PaginationParam
}

type ListPositionResponse struct {
	Data  []*Position
	Total int
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

func (uc *PositionUsecase) ListPosition(ctx context.Context, req *ListPositionRequest) (*ListPositionResponse, error) {
	return uc.repo.ListPosition(ctx, req)
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
