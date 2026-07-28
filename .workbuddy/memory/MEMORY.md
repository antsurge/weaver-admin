# weaver-admin 项目长期记忆

## 工具链与构建

### Go 版本管理（gvm）
- 系统 PATH 中 `go` 是 `gvm` 软链接指向 1.22.1，但项目 `go.mod` 用 `toolchain go1.24.11`
- 直接 `go build` 会用 1.22.1 编译，失败（缺 `iter` 等）
- **正确姿势**：绝对路径 + 强制 GOROOT
  ```bash
  GOROOT=/Users/wjf/.gvm/gos/go1.24.11 /Users/wjf/.gvm/gos/go1.24.11/bin/go build ./...
  ```
  或者 export PATH 后用 `GOTOOLCHAIN=auto`，但 GOROOT 仍可能指向旧版。绝对路径 + GOROOT 最稳。

### protoc 生成（替代 buf remote）
- buf remote plugin 因 TLS 证书问题拉不下来 → 改用本地 protoc + 手动 `--go_opt=Mproto.proto=path;package` 生成
- zsh `;` 被识别为命令分隔符 → M 参数必须包在单引号内：`'--go_opt=Mfoo.proto=path;package'`
- 推荐：`paths=source_relative`，用 shell for 循环批量生成所有 .proto 文件的 M 参数
- ent generate 需要 1.24+，故同 Go 版本管理

## 后端结构
- `app/admin/service/internal/{biz,data,service,handler,server,conf}` — 标准 kratos layout
- 数据模型：`biz.Menu` ↔ `data.Menu` ↔ `ent.Menu`，通过 `toBizMenu()` 互转
- m2m edge：必须双向声明，一侧 `edge.To`，另一侧 `edge.From(...).Ref(...)`，否则生成时报 missing inverse edge
- wire：手动改 `wire_gen.go` 是常见做法，不必每次跑 `go run wire/cmd/wire`
- 事务：`Data.InTx(ctx, fn)`，fn 里直接使用 ctx（业务层无需感知 TxClient）

## 前端约定
- 密码组件：`component: 'VbenInputPassword'`（不是 `'InputPassword'`）
- 类型字段：菜单 type 枚举定义在 `PermissionTypeOptionsValue` 常量里，值有 `catalog/menu/iframe/link/action`
- 表单模式："编辑时留空跳过"的字段（密码等），后端 builder 模式累加 SetXxx，最后判断是否调用 Set
- Sensitive 字段（ent `field.String("...").Sensitive()`）：JSON 序列化跳过，前端 Get API 不会收到，编辑表单对应字段保持空白

## 增量记忆
- 见 `2026-07-28.md` 里的具体任务（菜单接口权限、密码字段、openapi.yaml 路径修复等）

## openapi.yaml 路径解析
- `NewOpenAPIScanner` 在 `data.go` 中，接受 `*conf.OpenAPI` 参数
- `resolveOpenAPIPath(cfgPath)` 多路径 fallback：配置路径 → 项目根相对路径 → `../` → `../../` → Docker 内 `/app/` 路径
- 配置 `config.yaml` 中 `openapi.path: ""` 表示自动 fallback
- 所有路径都找不到时只记 warn 不阻塞启动（`Metadata()` 返回空 slice）
- Dockerfile 需额外 COPY `api/gen/openapi/openapi.yaml`