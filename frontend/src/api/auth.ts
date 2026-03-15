import type { LoginForm } from "./interface";
import { http } from "@/utils/http";

export function login(data: LoginForm) {
  return http.post<LoginForm>("/admin/v1/login", data, {});
}

export function getCaptcha() {
  return http.get("/admin/v1/get-captcha")
}