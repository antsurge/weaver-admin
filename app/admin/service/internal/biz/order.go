package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// Order 订单
type Order struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Sn        string     `json:"sn"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type ListOrderRequest struct {
	enthelper.PaginationParams
}

type ListOrderResponse struct {
	Data  []*Order
	Total int
}

type OrderRepo interface {
	List(context.Context, *ListOrderRequest) (*ListOrderResponse, error)
	Get(context.Context, string) (*Order, error)
	Create(context.Context, *Order) error
	Update(context.Context, *Order) error
	UpdateStatus(context.Context, string, string) error
	Delete(context.Context, []string) error
}

type OrderUsecase struct {
	repo OrderRepo
	log  *log.Helper
}

func NewOrderUsecase(repo OrderRepo, logger log.Logger) *OrderUsecase {
	return &OrderUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *OrderUsecase) List(ctx context.Context, req *ListOrderRequest) (*ListOrderResponse, error) {
	return uc.repo.List(ctx, req)
}
func (uc *OrderUsecase) Get(ctx context.Context, id string) (*Order, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *OrderUsecase) Create(ctx context.Context, req *Order) (*Order, error) {
	item := req
	item.ID = uuid.GenerateXID()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	err := uc.repo.Create(ctx, item)
	return item, err
}

func (uc *OrderUsecase) Update(ctx context.Context, req *Order) (*Order, error) {
	item := req
	item.UpdatedAt = time.Now()
	err := uc.repo.Update(ctx, item)
	return item, err
}
func (uc *OrderUsecase) UpdateStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}

func (uc *OrderUsecase) Delete(ctx context.Context, ids []string) error {
	return uc.repo.Delete(ctx, ids)
}
