package handler

import (
	"encoding/json"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/antsurge/weaver-admin/pkg/utils/crypto"
	"github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// ProfileHandler 个人中心相关的自服务 HTTP 处理。
// 仅允许修改当前登录用户自己的资料与密码，接口权限已豁免（见 http.go authz bypass）。
type ProfileHandler struct {
	adminUc *biz.AdminUseCase
}

func NewProfileHandler(adminUc *biz.AdminUseCase) *ProfileHandler {
	return &ProfileHandler{adminUc: adminUc}
}

// UpdateCurrentUser 修改当前登录用户的资料（昵称 realName / 头像 avatar）。
// PUT /admin/v1/current-user
// body: { "realName": "张三", "avatar": "uploads/avatar/xxx.png" }
func (h *ProfileHandler) UpdateCurrentUser(ctx khttp.Context) error {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return errors.Unauthorized("UNAUTHORIZED", "请先登录")
	}

	body := struct {
		RealName string `json:"realName"`
		Avatar   string `json:"avatar"`
	}{}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return errors.BadRequest("BAD_REQUEST", "请求体解析失败")
	}

	// 至少需要修改一项
	if body.RealName == "" && body.Avatar == "" {
		return errors.BadRequest("BAD_REQUEST", "昵称或头像不能同时为空")
	}

	// UpdateAdmin 会无条件写入 realName/username/email/phone/avatar/department
	// 等字段，空字符串会清空原值；因此先读取当前用户回填未修改字段，
	// 避免个人中心改名/换头像时误清空账号的其他信息。
	cur, err := h.adminUc.GetAdminWithRoles(ctx, adminID)
	if err != nil {
		return err
	}
	toUpdate := &biz.Admin{
		ID:           adminID,
		Username:     cur.Username,
		RealName:     cur.RealName,
		Email:        cur.Email,
		Phone:        cur.Phone,
		Avatar:       cur.Avatar,
		DepartmentID: cur.DepartmentID,
	}
	if body.RealName != "" {
		toUpdate.RealName = body.RealName
	}
	if body.Avatar != "" {
		toUpdate.Avatar = body.Avatar
	}

	if _, err := h.adminUc.UpdateAdmin(ctx, toUpdate); err != nil {
		return err
	}

	return ctx.Result(200, map[string]any{"updated": true})
}

// UpdateCurrentUserPassword 修改当前登录用户的密码（校验旧密码）。
// PUT /admin/v1/current-user/password
// body: { "oldPassword": "旧密码", "newPassword": "新密码" }
func (h *ProfileHandler) UpdateCurrentUserPassword(ctx khttp.Context) error {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return errors.Unauthorized("UNAUTHORIZED", "请先登录")
	}

	body := struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}{}
	if err := json.NewDecoder(ctx.Request().Body).Decode(&body); err != nil {
		return errors.BadRequest("BAD_REQUEST", "请求体解析失败")
	}

	if body.OldPassword == "" || body.NewPassword == "" {
		return errors.BadRequest("BAD_REQUEST", "旧密码与新密码不能为空")
	}
	if len(body.NewPassword) < biz.MinPasswordLength {
		return errors.BadRequest("PASSWORD_TOO_SHORT", "新密码长度不能少于6位")
	}

	admin, err := h.adminUc.GetAdminWithRoles(ctx, adminID)
	if err != nil {
		return err
	}

	// 校验旧密码（若当前密码为空，视为未设置密码，直接放行）
	if admin.Password != "" && !crypto.CheckPasswordHash(body.OldPassword, admin.Password) {
		return errors.BadRequest("OLD_PASSWORD_INCORRECT", "旧密码不正确")
	}

	if err := h.adminUc.ResetPassword(ctx, adminID, body.NewPassword); err != nil {
		return err
	}

	return ctx.Result(200, map[string]any{"updated": true})
}
