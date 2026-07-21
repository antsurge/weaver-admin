package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	identityV1 "github.com/antsurge/weaver-admin/api/gen/go/identity/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
)

type IdentityService struct {
	adminV1.UnimplementedIdentityServer

	adminUc *biz.AdminUseCase
}

func NewIdentityService(adminUc *biz.AdminUseCase) *IdentityService {
	return &IdentityService{adminUc: adminUc}
}

func (s *IdentityService) ListAdmin(ctx context.Context, req *identityV1.ListAdminRequest) (*identityV1.ListAdminResponse, error) {
	input := biz.ListAdminRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.adminUc.ListAdmin(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &identityV1.ListAdminResponse{}
	if res != nil {
		err = copierx.Copy(&output, res)
	}

	return output, nil
}

func (s *IdentityService) GetAdmin(ctx context.Context, req *identityV1.GetAdminRequest) (*identityV1.Admin, error) {
	input := biz.Admin{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	admin, err := s.adminUc.CreateAdmin(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := &identityV1.Admin{}
	err = copierx.Copy(output, admin)

	return output, err
}

func (s *IdentityService) CreateAdmin(ctx context.Context, req *identityV1.CreateAdminRequest) (*identityV1.Admin, error) {
	input := biz.Admin{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	admin, err := s.adminUc.CreateAdmin(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := &identityV1.Admin{}
	err = copierx.Copy(output, admin)

	return output, err
}

func (s *IdentityService) UpdateAdmin(ctx context.Context, req *identityV1.UpdateAdminRequest) (*identityV1.Admin, error) {
	input := biz.Admin{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	admin, err := s.adminUc.UpdateAdmin(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := &identityV1.Admin{}
	err = copierx.Copy(output, admin)

	return output, err
}
