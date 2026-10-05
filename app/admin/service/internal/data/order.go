package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/order"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.OrderRepo = (*orderRepo)(nil)

type orderRepo struct {
	data *Data
	log  *log.Helper
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{data: data, log: log.NewHelper(logger)}
}

func (r *orderRepo) List(ctx context.Context, params *biz.ListOrderRequest) (*biz.ListOrderResponse, error) {
	query := r.data.db.Order.Query().Order(ent.Desc(order.FieldCreatedAt)).
		Where(order.DeletedAtIsNil())

	res, err := enthelper.Pagination[*ent.Order, *ent.OrderQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.Order, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.Order{
			ID:        v.ID,
			Name:      v.Name,
			Sn:        v.Sn,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			DeletedAt: v.DeletedAt,
		})
	}
	return &biz.ListOrderResponse{Data: data, Total: res.Total}, nil
}
func (r *orderRepo) Get(ctx context.Context, id string) (*biz.Order, error) {
	v, err := r.data.db.Order.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &biz.Order{
		ID:        v.ID,
		Name:      v.Name,
		Sn:        v.Sn,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		DeletedAt: v.DeletedAt,
	}, nil
}

func (r *orderRepo) Create(ctx context.Context, d *biz.Order) error {
	_, err := r.data.db.Order.Create().
		SetID(d.ID).
		SetName(d.Name).
		SetSn(d.Sn).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *orderRepo) Update(ctx context.Context, d *biz.Order) error {
	_, err := r.data.db.Order.UpdateOneID(d.ID).
		SetName(d.Name).
		SetSn(d.Sn).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}
func (r *orderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Order.UpdateOneID(id).
		SetStatus(order.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *orderRepo) Delete(ctx context.Context, ids []string) error {
	now := time.Now()
	_, err := r.data.db.Order.Update().
		Where(order.IDIn(ids...)).
		SetDeletedAt(now).
		Save(ctx)
	return err
}
