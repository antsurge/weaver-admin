package pb

import (
	"context"

	v1 "github.com/antsurge/weaver-admin/api/gen/go/organization/service/v1"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
)

var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationOrganzationExportPosition = "/admin.service.v1.Ogranization/ExportPosition"
const OperationOrganzationImportPosition = "/admin.service.v1.Ogranization/ImportPosition"
const OperationOrganzationDownloadPositionTemplate = "/admin.service.v1.Ogranization/DownloadPositionTemplate"

type OrganizationHandlerHTTPServer interface {
	ExportPosition(http.Context) error
	ImportPosition(http.Context) error
	DownloadPositionTemplate(http.Context) error
}

func RegisterOrganizationHandlerServer(s *http.Server, srv OrganizationHandlerHTTPServer) {
	r := s.Route("/")
	r.POST("admin/v1/position:export", _Organization_ExportPosition0_HTTP_Handler(srv))
	r.POST("admin/v1/position:import", _Organization_ImportPosition0_HTTP_Handler(srv))
	r.GET("admin/v1/position:template", _Organization_DownloadPositionTemplate0_HTTP_Handler(srv))
}

func _Organization_ExportPosition0_HTTP_Handler(srv OrganizationHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in v1.ListPositionRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationOrganzationExportPosition)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
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
			mergeContext(ctx, ctx1)
			return nil, srv.ImportPosition(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

func _Organization_DownloadPositionTemplate0_HTTP_Handler(srv OrganizationHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in v1.ListPositionRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationOrganzationDownloadPositionTemplate)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.DownloadPositionTemplate(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}
