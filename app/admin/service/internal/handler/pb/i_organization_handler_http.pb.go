package pb

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
	v1 "github.com/hypercoze/kratos-admin/api/gen/go/organization/service/v1"
)

var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationOrganzationExportPosition = "/admin.service.v1.Ogranization/ExportPosition"
const OperationOrganzationImportPosition = "/admin.service.v1.Ogranization/ImportPosition"

type OrganizationHandlerHTTPServer interface {
	ExportPosition(http.Context) error
	ImportPosition(http.Context) error
}

func RegisterOrganizationHandlerServer(s *http.Server, srv OrganizationHandlerHTTPServer) {
	r := s.Route("/")
	r.POST("admin/v1/position:export", _Organization_ExportPosition0_HTTP_Handler(srv))
	r.POST("admin/v1/position:import", _Organization_ImportPosition0_HTTP_Handler(srv))
}

func _Organization_ExportPosition0_HTTP_Handler(srv OrganizationHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in v1.ListPositionRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationOrganzationExportPosition)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			return nil, srv.ExportPosition(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

func _Organization_ImportPosition0_HTTP_Handler(srv OrganizationHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in v1.ListPositionRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationOrganzationImportPosition)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			return nil, srv.ImportPosition(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}
