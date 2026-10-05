package service

import (
	"context"

	securityV1 "github.com/antsurge/weaver-admin/api/gen/go/security/service/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListOnlineUsers 在线用户列表（安全中心-在线用户）
func (s *SecurityService) ListOnlineUsers(ctx context.Context, _ *emptypb.Empty) (*securityV1.ListOnlineUsersResponse, error) {
	items, err := s.onlineUc.ListOnlineUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := &securityV1.ListOnlineUsersResponse{Total: int64(len(items))}
	out.Items = make([]*securityV1.OnlineUser, 0, len(items))
	for _, it := range items {
		out.Items = append(out.Items, &securityV1.OnlineUser{
			Id:           it.ID,
			Username:     it.Username,
			RealName:     it.RealName,
			Ip:           it.IP,
			UserAgent:    it.UserAgent,
			LoginAt:      it.LoginAt.Unix(),
			ExpireAt:     it.ExpireAt.Unix(),
			SessionCount: int32(it.SessionCount),
		})
	}
	return out, nil
}

// ForceLogout 强制下线（安全中心-在线用户）
func (s *SecurityService) ForceLogout(ctx context.Context, req *securityV1.ForceLogoutRequest) (*emptypb.Empty, error) {
	if err := s.onlineUc.ForceLogout(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
