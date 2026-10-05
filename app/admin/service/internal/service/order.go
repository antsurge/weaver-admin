package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	pb "github.com/antsurge/weaver-admin/api/gen/go/order/service/v1"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

// OrderService 订单服务
type OrderService struct {
	adminV1.UnimplementedOrderServiceServer
	uc  *biz.OrderUsecase
	log *log.Helper
}

func NewOrderService(uc *biz.OrderUsecase, logger log.Logger) *OrderService {
	return &OrderService{uc: uc, log: log.NewHelper(logger)}
}

func (s *OrderService) ListOrder(ctx context.Context, req *pb.ListOrderRequest) (*pb.ListOrderResponse, error) {
	res, err := s.uc.List(ctx, &biz.ListOrderRequest{
		PaginationParams: enthelper.PaginationParams{CurrentPage: int(req.GetCurrentPage()), PageSize: int(req.GetPageSize())},
	})
	if err != nil {
		return nil, err
	}
	items := make([]*pb.Order, 0, res.Total)
	for _, v := range res.Data {
		items = append(items, &pb.Order{
			Id:   v.ID,
			Name: v.Name,
			Sn:   v.Sn,
		})
	}
	return &pb.ListOrderResponse{Items: items, Total: int64(res.Total)}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	item, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetOrderResponse{Data: &pb.Order{
		Id:   item.ID,
		Name: item.Name,
		Sn:   item.Sn,
	}}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	item := &biz.Order{
		Name: req.GetName(),
		Sn:   req.GetSn(),
	}
	created, err := s.uc.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrderResponse{Data: &pb.Order{Id: created.ID}}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	item := &biz.Order{
		ID:   req.GetId(),
		Name: req.GetName(),
		Sn:   req.GetSn(),
	}
	updated, err := s.uc.Update(ctx, item)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateOrderResponse{Data: &pb.Order{Id: updated.ID}}, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	if err := s.uc.Delete(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &pb.DeleteOrderResponse{}, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	if err := s.uc.UpdateStatus(ctx, req.GetId(), req.GetStatus()); err != nil {
		return nil, err
	}
	return &pb.UpdateOrderStatusResponse{}, nil
}
