package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	tenantV1 "github.com/antsurge/weaver-admin/api/gen/go/tenant/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TenantService struct {
	adminV1.UnimplementedTenantServer

	tenantUc *biz.TenantUsecase
}

func NewTenantService(tenantUc *biz.TenantUsecase) *TenantService {
	return &TenantService{tenantUc: tenantUc}
}

func (s *TenantService) ListTenant(ctx context.Context, req *tenantV1.ListTenantRequest) (*tenantV1.ListTenantResponse, error) {
	input := biz.ListTenantRequest{}
	if err := copier.Copy(&input, req); err != nil {
		return nil, err
	}

	res, err := s.tenantUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &tenantV1.ListTenantResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		if err = copierx.Copy(&output.Items, &res.Data); err != nil {
			return nil, err
		}
	}
	return output, nil
}

func (s *TenantService) GetTenant(ctx context.Context, req *tenantV1.GetTenantRequest) (*tenantV1.Tenant, error) {
	t, err := s.tenantUc.GetTenant(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	output := &tenantV1.Tenant{}
	if err = copier.Copy(output, t); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *TenantService) CreateTenant(ctx context.Context, req *tenantV1.CreateTenantRequest) (*tenantV1.Tenant, error) {
	input := biz.Tenant{}
	if err := copier.Copy(&input, req); err != nil {
		return nil, err
	}

	t, err := s.tenantUc.CreateTenant(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &tenantV1.Tenant{}
	if err = copier.Copy(output, t); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *TenantService) UpdateTenant(ctx context.Context, req *tenantV1.UpdateTenantRequest) (*tenantV1.Tenant, error) {
	input := biz.Tenant{ID: req.Id}
	if err := copier.Copy(&input, req); err != nil {
		return nil, err
	}

	t, err := s.tenantUc.UpdateTenant(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &tenantV1.Tenant{}
	if err = copier.Copy(output, t); err != nil {
		return nil, err
	}
	return output, nil
}

func (s *TenantService) UpdateTenantStatus(ctx context.Context, req *tenantV1.UpdateTenantStatusRequest) (*tenantV1.Tenant, error) {
	if err := s.tenantUc.UpdateTenantStatus(ctx, req.Id, req.Status); err != nil {
		return nil, err
	}
	// 返回更新后的租户详情
	return s.GetTenant(ctx, &tenantV1.GetTenantRequest{Id: req.Id})
}

func (s *TenantService) DeleteTenant(ctx context.Context, req *tenantV1.DeleteTenantRequest) (*emptypb.Empty, error) {
	if err := s.tenantUc.DeleteTenant(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
