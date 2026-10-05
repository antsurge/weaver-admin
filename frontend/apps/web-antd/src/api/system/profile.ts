import { requestClient } from '#/api/request';

/**
 * 修改当前用户资料（个人中心自服务接口，无需接口权限）
 * @param data 修改资料参数
 * @param data.avatar 头像对象键
 * @param data.realName 昵称
 */
async function updateCurrentUserApi(data: {
  /** 头像对象键，如 uploads/avatar/xxx.png（可选） */
  avatar?: string;
  /** 昵称（可选） */
  realName?: string;
}) {
  return requestClient.put('/admin/v1/current-user', data, {
    showSuccessMessage: true,
  });
}

/**
 * 修改当前用户密码（校验旧密码）
 * @param data 修改密码参数
 * @param data.newPassword 新密码
 * @param data.oldPassword 旧密码
 */
async function updateCurrentUserPasswordApi(data: {
  newPassword: string;
  oldPassword: string;
}) {
  return requestClient.put('/admin/v1/current-user/password', data, {
    showSuccessMessage: true,
  });
}

export { updateCurrentUserApi, updateCurrentUserPasswordApi };
