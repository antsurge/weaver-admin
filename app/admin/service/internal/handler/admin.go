package handler

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// AdminHandler 用户管理相关的自定义 HTTP 处理（导出等）
type AdminHandler struct {
	adminUc *biz.AdminUseCase
}

func NewAdminHandler(adminUc *biz.AdminUseCase) *AdminHandler {
	return &AdminHandler{adminUc: adminUc}
}

// ExportAdmin 导出用户列表为 CSV（UTF-8 BOM，Excel 可直接打开）
func (h *AdminHandler) ExportAdmin(ctx http.Context) error {
	req := ctx.Request()
	q := req.URL.Query()

	params := &biz.ListAdminRequest{}
	params.CurrentPage = 1
	// 导出不分页：PageSize=0 表示取全部记录（见 enthelper.Pagination）
	params.PageSize = 0
	params.Username = q.Get("username")
	params.RealName = q.Get("realName")
	params.Phone = q.Get("phone")
	params.Email = q.Get("email")
	params.Status = q.Get("status")

	res, err := h.adminUc.ListAdmin(ctx, params)
	if err != nil {
		return err
	}

	filename := "admin_" + time.Now().Format("20060102150405") + ".csv"
	// 中文文件名需 URL 编码
	ctx.Response().Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
			filename, url.PathEscape("用户列表_"+time.Now().Format("20060102150405")+".csv")),
	)
	ctx.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	ctx.Response().WriteHeader(200)

	// 写入 UTF-8 BOM，避免 Excel 打开中文乱码
	if _, err := ctx.Response().Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}

	w := csv.NewWriter(ctx.Response())
	_ = w.Write([]string{"用户名", "姓名", "手机号", "邮箱", "角色", "状态", "创建时间"})
	for _, item := range res.Items {
		status := "启用"
		if item.Status == biz.StatusDisabled {
			status = "禁用"
		}
		_ = w.Write([]string{
			item.Username,
			item.RealName,
			item.Phone,
			item.Email,
			joinStrings(item.RoleNames),
			status,
			item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
	return w.Error()
}

func joinStrings(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for _, s := range items[1:] {
		result += "、" + s
	}
	return result
}
