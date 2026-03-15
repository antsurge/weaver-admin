import { StatusCodes, getReasonPhrase } from "http-status-codes";

export function handleHttpStatus(status: number) {
  switch (status) {
    case StatusCodes.UNAUTHORIZED:
      console.error("未登录");
      break;

    case StatusCodes.FORBIDDEN:
      console.error("无权限");
      break;

    case StatusCodes.NOT_FOUND:
      console.error("接口不存在");
      break;

    case StatusCodes.INTERNAL_SERVER_ERROR:
      console.error("服务器错误");
      break;

    default:
      console.error(getReasonPhrase(status));
  }
}
