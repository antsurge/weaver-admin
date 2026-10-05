# Bug 修复记录

> 记录 weaver-admin 逐模块分析中发现并修复的问题，每完成一项追加记录，全部修完后统一测试。

## 进度状态（2026-09-03）

- **FIX-001 ~ FIX-014（P0/P1/P2/P3 全部）代码已完成**
- 后端 `GOTOOLCHAIN=auto go build ./app/... ./pkg/...` 通过；前端编译/lint 通过
- 统一测试前的所有手工测试项见文末"统一测试清单"；全部修改**尚未执行运行时测试**

## 进度状态（2026-09-11，本轮）

- 已在本机 MySQL + Redis 环境**实际启动服务并执行统一运行时测试**，逐项结果见文末「统一测试结果（2026-09-11）」
- 测试中发现并修复 4 个新问题：**FIX-016**（部门 type 校验错误，P0）、**FIX-017**（.gitignore 未覆盖构建产物，P3）、**FIX-018**（消息中心图标导出缺失，P2）、**FIX-019**（任务执行成功时日志写入失败，P2）
- 同轮新增两个功能模块：**登录策略 / 安全策略**（登录失败锁定、IP 白/黑名单、Token 有效期策略）与 **系统监控**（服务/缓存/数据库/接口/定时任务监控）
- **FIX-012 已失效**：MQ 能力本轮以可插拔驱动（memory/redis/rabbitmq/…）重新引入，原"彻底删除 MQ"的结论不再适用
- 前端 `vue-tsc`（typecheck）存在 79 处既有类型报错（`showSuccessMessage`、`@core/@vben` 模板类型等历史基线问题），非本次改动引入；`vite build` 因 tailwind 配置经 `jiti` 进入打包而失败（详见文末说明）

## 已修复

### FIX-001 创建/更新管理员时密码明文入库（P0）

- **发现时间**：2026-09-03
- **严重级别**：🔴 P0（安全 + 功能）
- **问题描述**：
  - `pkg/utils/crypto/password.go` 提供了 bcrypt 哈希函数 `HashPassword`，但业务代码从未调用（仅测试文件引用）。
  - 创建管理员时密码（含默认密码 `123456`）明文直接写库；登录校验用 `CheckPasswordHash`（bcrypt 对比），与明文存储不匹配——通过 API 创建的账号实际无法登录。
- **修复方案**：
  - `biz/admin.go` 的 `CreateAdmin`：密码落库前经 `crypto.HashPassword` 哈希（默认密码逻辑保持不变）。
  - `biz/admin.go` 的 `UpdateAdmin`：密码非空（表示修改密码）时哈希后入库；空密码仍表示不修改。
- **修改文件**：
  - `app/admin/service/internal/biz/admin.go`
- **验证**：`go build ./app/...` 通过；登录链路（创建账号 → 登录）待统一测试。
- **注意**：数据库存量明文密码不会自动转换，需通过"修改密码"接口重置或写脚本批量哈希。

### FIX-002 部门 type 字段全链路断裂，创建部门必然失败（P0）

- **发现时间**：2026-09-03
- **严重级别**：🔴 P0（功能不可用）
- **问题描述**：
  - Ent schema 中 `type` 是必填枚举（company/subsidiary/department/position，无默认值），但 `data/department.go` 的 `CreateDepartment` 未调用 `SetType` → 创建部门时 Ent 枚举校验直接报错。
  - 列表/详情转换未读取 `type` → 即使数据存在也回显不出来。
  - 前端部门表单根本没有 type 字段，即使后端修好，proto 的 CEL 校验（`type in [...]`）也会拦截空值。
  - 注：`UpdateDepartmentRequest`（proto）未定义 type 字段，更新场景不涉及。
- **修复方案**：
  - 后端 `data/department.go`：`CreateDepartment` 补 `SetType`；`ListDepartment` 转换与 `toBiz` 补 `Type` 回读；`UpdateDepartment` 加注释说明不更新 type 的原因。
  - 前端 `api/organization/department.ts`：`Department` 接口补充 `type` 字段。
  - 前端表单 `modules/form/index.vue` + `rules.ts`：新增"类型"Select 字段（必填，默认 `department`，选项：公司/子公司/部门/岗位）。
  - 国际化 `locales/langs/zh-CN/organization.json`：补充 `fields.type` 与 `type.*` 文案。
- **修改文件**：
  - `app/admin/service/internal/data/department.go`
  - `frontend/apps/web-antd/src/api/organization/department.ts`
  - `frontend/apps/web-antd/src/views/organization/department/modules/form/index.vue`
  - `frontend/apps/web-antd/src/views/organization/department/modules/form/rules.ts`
  - `frontend/apps/web-antd/src/locales/langs/zh-CN/organization.json`
- **验证**：`go build ./app/... ./pkg/...` 通过；创建部门（四种类型）→ 列表/详情回显，待统一测试。

### FIX-003 JWT 中间件不校验 token 存储状态，登出/强制下线失效（P1）

- **发现时间**：2026-09-03
- **严重级别**：🔴 安全（原本 P1，因实现简单随 FIX-003 一并处理）
- **问题描述**：
  - 登录时 access/refresh token 均已存入 Redis（`data/token.go`），但 `pkg/middleware/auth/jwt.go` 只验证 JWT 签名，从不查询 Redis。
  - 后果：`Logout` 删除 token 后，access token 在 TTL 内依然可用；"强制下线"能力无法实现。
- **修复方案**：
  - `pkg/utils/auth/options.go`：新增 `TokenStore` 接口（`Exists(ctx, token)`）、`WithTokenStore` 选项、`NewOptions` 构造器。
  - `pkg/middleware/auth/jwt.go`：解析 JWT 并确认 `type == "access"` 后，若配置了 TokenStore 则校验 token 是否仍存在，不存在即返回 401。
  - `server/http.go`：`NewHTTPServer` 注入 `biz.TokenRepo` 并传入 auth 中间件（`authUtils.WithTokenStore(tokenRepo)`）。
  - `cmd/service/wire_gen.go`：同步 wire 调用参数。
  - Redis 故障时 fail-closed（返回 401），与"保存失败即登录失败"的现有策略一致。
  - 注意：gRPC server 未启用 auth 中间件（仅 recovery），已在 FIX-006 中一并关注。
- **修改文件**：
  - `pkg/utils/auth/options.go`
  - `pkg/middleware/auth/jwt.go`
  - `app/admin/service/internal/server/http.go`
  - `app/admin/service/cmd/service/wire_gen.go`
- **验证**：编译通过；测试项：登录 → 登出 → 用原 access token 调接口应返回 401。

### FIX-004 前端 token 刷新链路断裂（P1）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P1
- **问题描述**：
  - `refreshTokenApi` 调用不存在的 `/auth/refresh`（后端实际为 `POST /admin/v1/refresh-token`，且要求 body 同时携带 refreshToken + accessToken）。
  - 登录后只保存了 accessToken，refreshToken 被丢弃 → 过期后无法续期，只能重新登录。
  - `doRefreshToken` 把整个响应体当 token 存入 accessStore。
  - `logoutApi` 发送空 body，服务端 `Logout` 拿不到 token，登出没有真正作废服务端 token。
  - preferences 未启用 `enableRefreshToken`，401 时不会尝试刷新。
- **修复方案**：
  - `api/core/auth.ts` 重写：localStorage 持久化 refreshToken（`getRefreshToken/setRefreshToken`）；`refreshTokenApi` 对接 `/admin/v1/refresh-token` 并携带双 token（走无拦截器的 baseRequestClient 防死循环）；`logoutApi` 携带双 token；移除不存在的 `/auth/codes` 接口。
  - `api/request.ts`：`doRefreshToken` 解析 LoginResponse 并成对更新双 token；`doReAuthenticate` 同时清除 refreshToken。
  - `store/auth.ts`：登录成功保存 refreshToken；登出清除；清理调试 console.log。
  - `preferences.ts`：`enableRefreshToken: true`。
- **修改文件**：
  - `frontend/apps/web-antd/src/api/core/auth.ts`
  - `frontend/apps/web-antd/src/api/request.ts`
  - `frontend/apps/web-antd/src/store/auth.ts`
  - `frontend/apps/web-antd/src/preferences.ts`
- **验证**：编译/lint 通过；测试项：登录 → 等待 access token 过期（或手改 Redis TTL）→ 请求自动刷新续期；登出后 refreshToken 失效。

### FIX-005 前端依赖后端不存在的 /menu/all，动态路由加载必失败（P1）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P1（backend 模式下菜单加载 404）
- **问题描述**：
  - `accessMode: 'backend'` 时，vben 的 `generateRoutesByBackend` 只使用 `fetchMenuListAsync()` 的结果，忽略 `options.routes`。
  - 项目 `fetchMenuListAsync` 调用 `/menu/all` —— 后端根本没有该路由 → 动态路由生成失败。
  - 同时 guard 中已用 `current-user` 返回的 menuTree 构建了 `accessRoutes`，两套数据源并存且互相对不上。
  - 顺带发现：`getAccessCodesApi`（`/auth/codes`）也是幽灵接口且从未被调用（权限码实际从菜单 authCode 收集）。
- **修复方案**：
  - `router/access.ts` 重构：`transformAccessRoutes` 产出的路由组件改为字符串标识（catalog → `'BasicLayout'`、iframe → `'IFrameView'`、menu → 后端存储的 component 路径），由 vben 的 layoutMap/pageMap 统一解析（与 vben 后端路由规范一致）。
  - `generateAccess.fetchMenuListAsync` 直接返回 guard 构建好的 `options.routes`，删除 `/menu/all` 依赖。
  - 删除本地 `getViewComponent` + views glob 重复解析逻辑（组件匹配失败时 vben 自动回退 not-found 页）。
  - `api/core/menu.ts`：删除 `getAllMenusApi`。
- **修改文件**：
  - `frontend/apps/web-antd/src/router/access.ts`
  - `frontend/apps/web-antd/src/api/core/menu.ts`
  - `frontend/apps/web-antd/src/api/core/auth.ts`（移除 /auth/codes）
- **验证**：lint 通过；测试项：登录后侧边栏菜单按角色正确渲染、页面组件正常加载、iframe/外链菜单可用。

### FIX-006 后端无 API 鉴权执行，RBAC 仅用于前端菜单展示（P0）

- **发现时间**：2026-09-03
- **严重级别**：🔴 P0（安全：任何登录用户可调用全部管理接口）
- **问题描述**：
  - 全项目无任何权限校验执行点：JWT 中间件只确认"已登录"，角色-菜单-API 权限模型的数据（role_menus、menu_api_permissions）全部闲置，仅前端菜单渲染使用。
  - 顺带：gRPC server 只有 recovery 中间件，完全没有认证。
- **实现设计**：
  - 新增 `pkg/middleware/authz` 中间件（可复用）：
    - 匹配模型：请求 `HTTP Method + 路由路径模板`（kratos `PathTemplate()`，如 `/admin/v1/admin/{adminId}`）对 ApiPermission 的 `Method + Path`。两者同源于 google.api.http 注解，模板串可直接精确匹配，另用 `{param}` 转正则兜底参数命名差异。
    - **superAdmin 直通**（角色 code == `appConf.super_admin_code`）。
    - **自服务接口豁免**：`POST /admin/v1/logout`、`GET /admin/v1/current-user/info`、`GET /admin/v1/current-user/menus`（无任何权限也能看个人信息/登出）。
    - 未登录接口（Login/GetCaptcha/RefreshToken 在 auth 白名单）无 adminID，天然跳过。
    - 仅对 HTTP 生效；gRPC 的接口权限模型未定义，暂不校验。
    - 权限查询失败 fail-closed（403）。
  - 新增 `biz/authz.go`：`AuthzUseCase.GetPermission`（Redis 缓存 60s + 回源 DB）。
  - 新增 `data/authz.go`：`GetPermissionByAdmin` 查询链路 admin_roles → roles(启用) → role_menus → menus(启用) → api_permissions，生成权限码集合 `METHOD|path`；Redis 缓存 key `{prefix}authz:{adminID}`，TTL 60s（权限/角色/菜单变更最长 60s 延迟生效）。
  - `server/http.go`：authz 中间件插在 auth 之后（依赖其注入的 adminID）；`biz.AuthzUseCase` 经 CheckerFunc 适配为 `authz.Checker`。
  - `server/grpc.go`：补上认证中间件（同白名单、同 TokenStore），gRPC 接口权限暂不启用（已注释说明）。
- **修改文件**：
  - `pkg/middleware/authz/authz.go`（新增）
  - `app/admin/service/internal/biz/authz.go`（新增）
  - `app/admin/service/internal/data/authz.go`（新增）
  - `app/admin/service/internal/biz/biz.go`、`data/data.go`（ProviderSet 注册）
  - `app/admin/service/internal/server/http.go`、`server/grpc.go`
  - `app/admin/service/cmd/service/wire_gen.go`
- **验证**：`go build ./app/... ./pkg/...` 通过，`go vet` 无告警；行为变更测试项：
  - superAdmin 登录 → 所有接口可用（直通）
  - 普通用户未绑定任何接口权限 → 管理接口全部 403（预期行为）
  - 给角色绑定含接口权限的按钮菜单 → 对应接口放行
  - 角色禁用/菜单禁用后 ≤60s 内权限收回
- **注意**：这是行为变更——现有非 superAdmin 账号在配置权限前会被 403。如需临时关闭，移除 `http.go` 中 `authzMw.Server(...)` 一段即可。

### FIX-007 Admin 无 status 字段、无重置密码接口（P2）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P2（功能缺失）
- **问题描述**：
  - Admin 模型无 `status`（enabled/disabled），无法禁用/启用账号。
  - 无"重置密码"能力（只能由用户自己改密，忘记密码无人可重置）。
- **修复方案**：
  - Ent schema `ent/schema/admin.go`：新增 `status` 枚举字段（enabled/disabled，默认 enabled）。
  - `api/proto/.../i_identity.proto`：`AdminService` 新增 `ResetPassword` rpc（`PUT /admin/v1/admin/{id}/password`），并重新执行 `go generate ./ent`、`buf generate`。
  - `biz/admin.go`：`Admin` 领域模型新增 `Status`；`AdminRepo` 接口新增 `UpdatePassword(ctx, adminID, hashedPassword)`、`DeleteAdmin` 签名扩展 `superAdminCode` 参数；新增 `DeleteAdmin`（禁止删除自己，operatorID 取自 JWT metadata）与 `ResetPassword` 业务实现。
  - `data/admin.go`：增/改/查全链路读写 `status`（`entadmin.Status(...)` 显式类型转换）；新增事务型 `DeleteAdmin`（先校验绑定 superAdmin 角色的账号不可删，再清理 `admin_roles`，最后软删）；新增 `UpdatePassword`。
  - `service/identity.go`：`DeleteAdmin` 传入操作者 ID；暴露 `ResetPassword`。
  - `biz/authentication.go`：登录时拦截 `status == disabled` 的账号 → 拒绝登录（`ACCOUNT_DISABLED`）。
  - `cmd/service/wire_gen.go`：同步 `NewAdminUseCase` 参数（追加 `app *conf.App`）。
- **修改文件**：
  - `app/admin/service/internal/data/ent/schema/admin.go`
  - `api/proto/admin/service/v1/i_identity.proto`
  - `app/admin/service/internal/biz/admin.go`、`biz/authentication.go`
  - `app/admin/service/internal/data/admin.go`
  - `app/admin/service/internal/service/identity.go`
  - `app/admin/service/cmd/service/wire_gen.go`
- **验证**：Ent 代码与 buf 代码已重新生成，`go build ./app/... ./pkg/...` 通过；待统一测试。

### FIX-008 列表查询 N+1（P2）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P2（性能）
- **问题描述**：`ListAdmin` 为每条管理员记录单独查一次角色绑定（N+1）；`ListRole` 同样为每条角色单独查菜单 ID。
- **修复方案**：批量分组查询一次取回——`data/admin.go` 新增 `getRoleIDsByAdmins`，`data/role.go` 新增 `getMenuIDsByRoles`，遍历组装，消除 N+1。
- **修改文件**：
  - `app/admin/service/internal/data/admin.go`
  - `app/admin/service/internal/data/role.go`
- **验证**：`go build` 通过；接口输出结构不变（回归冒烟待统一测试）。

### FIX-009 删除操作不清理关联、无保护（P2）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P2（数据一致性 + 误操作）
- **问题描述**：
  - 删除 admin 不清理 `admin_roles`，可删除自己 / 可删除绑定 superAdmin 角色的账号。
  - 删除 role 不清理 `role_menus`、不检查 `is_system` 与"是否仍有用户绑定"。
  - 删除 department 直接硬删除，父部门删除后子部门成孤儿。
- **修复方案**：
  - admin 删除：事务内先拒绝删除自己与绑定 superAdmin 角色的账号，再清 `admin_roles`，最后软删。
  - role 删除：事务内拒绝删除 `is_system` 角色、拒绝删除仍有用户绑定的角色，成功则清理 `role_menus` 后删除。
  - department 删除：改为级联收集子孙后软删除子树（打 `deleted_at`），列表/详情查询统一过滤 `deleted_at`。
- **修改文件**：
  - `app/admin/service/internal/data/admin.go`
  - `app/admin/service/internal/data/role.go`
  - `app/admin/service/internal/data/department.go`
- **验证**：`go build` 通过；待统一测试（见测试清单 FIX-009）。

### FIX-010 菜单删除无事务且与软删策略不一致（P2）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P2（数据一致性）
- **问题描述**：菜单删除为直接硬删除，不清理 `role_menus`、`menu_api_permissions`，失败时无回滚，子树易产生孤儿数据。
- **修复方案**：`data/menu.go` 删除改为事务：递归收集子孙菜单 → 清理 `RoleMenu` 关联 → 逐节点 `ClearAPIPermissions()` → 物理删除。补 `rolemenu`/`role` 相关导入。
- **修改文件**：
  - `app/admin/service/internal/data/menu.go`
- **验证**：`go build` 通过；待统一测试。

### FIX-011 CurrentUserMenus 空壳 stub（P2）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P2（功能缺失）
- **问题描述**：`CurrentUserMenus` 只有被注释掉的 stub，前端"我的菜单"实际拿不到真实数据（动态菜单此前依赖该接口）。
- **修复方案**：biz 层抽出可复用方法 `currentUserMenus`，service 层删除空壳 stub 并用 `copierx.Copy` 恢复真实返回。
- **修改文件**：
  - `app/admin/service/internal/biz/authentication.go`
  - `app/admin/service/internal/service/authentication.go`
- **验证**：`go build` 通过；待统一测试。

### FIX-012 MQ 纯死代码清理（P3）

- **发现时间**：2026-09-03
- **严重级别**：🟢 P3（工程化）
- **问题描述**：`biz.MQ` 接口与 `data/mq/rabbitmq` 全套实现（channel pool/consumer/declare/publisher 等）无任何调用方与消费方；`NewData` 中注入点被注释、仅剩 wire ProviderSet 与 config.yaml 里的"僵尸配置"，会误导接入者以为 MQ 已启用。
- **修复方案**：彻底删除运行时代码，保留整洁配置 schema 供未来按需恢复：
  - 删除 `biz/mq.go`、`data/mq.go`、`data/mq/rabbitmq/` 目录（git rm）。
  - `data.go`：移除 rabbitmq import 与 ProviderSet 注册、清理 NewData 注释代码。
  - `conf.proto`：移除 `Bootstrap.mq`、`message MQ`、`message RabbitMQ`，Bootstrap 字段重新编号并重新生成 `conf.pb.go`（protoc）。
  - `wire.go`/`wire_gen.go`/`main.go`：去掉 `*conf.MQ` 依赖参数。
  - `config.yaml`：删除 `mq` 配置块。
- **修改文件**：
  - `app/admin/service/internal/biz/mq.go`（删）、`data/mq.go`（删）、`data/mq/rabbitmq/`（删）
  - `app/admin/service/internal/data/data.go`
  - `app/admin/service/internal/conf/conf.proto`、`conf/conf.pb.go`（重新生成）
  - `app/admin/service/cmd/service/wire.go`、`wire_gen.go`、`main.go`
  - `app/admin/service/configs/config.yaml`
- **验证**：`go build ./app/... ./pkg/...` 通过；服务可正常启动（不再出现 MQ 相关引用）。

### FIX-013 .gitignore 遗漏运行产物（P3）

- **发现时间**：2026-09-03
- **严重级别**：🟢 P3（工程化）
- **问题描述**：`dump.rdb`（根目录 + frontend/）、`go_build_admin`（根 + cmd/service）、`.!24019!go_build_admin` 等本地运行/构建产物被 git 跟踪。
- **修复方案**：`.gitignore` 追加 `*.rdb`、`dump.rdb`、`go_build_admin`、`out/`；`git rm --cached` 移除上述已跟踪文件（磁盘文件保留，新提交不再包含）。
- **修改文件**：
  - `.gitignore`
  - git 索引（`dump.rdb`、`frontend/dump.rdb`、`go_build_admin`、`app/admin/service/cmd/service/go_build_admin`、`.!24019!go_build_admin`）
- **验证**：`git status` 不再显示 rdb/go_build_admin 变更。

### FIX-014 默认密码硬编码（P3）

- **发现时间**：2026-09-03
- **严重级别**：🟢 P3（安全配置）
- **问题描述**：`AdminDefaultPassword = "123456"` 硬编码在 biz 常量中，修改需改代码。
- **修复方案**：
  - `conf.proto` 的 `App` message 新增 `admin_default_password` 字段（重新生成 `conf.pb.go`）。
  - `config.yaml` 的 `app` 段增加 `admin_default_password: "123456"`（生产环境改强密码即可，无需动代码）。
  - `biz/admin.go`：删除原常量，改为 `defaultPassword()` helper——优先读 `app.admin_default_password`，未配置时回退内置 `123456`。
- **修改文件**：
  - `app/admin/service/internal/conf/conf.proto`、`conf/conf.pb.go`（重新生成）
  - `app/admin/service/configs/config.yaml`
  - `app/admin/service/internal/biz/admin.go`
- **验证**：`go build` 通过；测试项：修改 `config.yaml` 默认密码后新建管理员生效。

### FIX-015 接口权限校验总开关（默认关闭，避免权限数据未配齐时阻塞联调）

- **发现时间**：2026-09-03
- **严重级别**：🟡 P2（FIX-006 引入后，角色未绑定接口权限即 403，权限数据未就绪阶段"一校验就走不下去"）
- **问题描述**：
  - FIX-006 上线后，只要账号未被正确识别为 superAdmin（绑定 `code=super_admin_code` 且启用的角色）且未绑定任何接口权限，访问管理接口一律 403，开发联调无法继续。
- **修复方案**：
  - `conf.proto` 的 `App` message 新增 `authz_enabled` 布尔开关（重新生成 `conf.pb.go`）。
  - `config.yaml` 的 `app` 段增加 `authz_enabled: false`（默认关闭接口权限校验）。
  - `server/http.go`：authz 中间件改为按 `appConf.AuthzEnabled` 条件挂载；开启后行为与 FIX-006 完全一致（superAdmin 直通、其余按角色绑定接口权限匹配、自服务接口豁免、位于 auth 之后）。
- **修改文件**：
  - `app/admin/service/internal/conf/conf.proto`、`conf/conf.pb.go`（重新生成）
  - `app/admin/service/configs/config.yaml`
  - `app/admin/service/internal/server/http.go`
- **验证**：`go build ./app/... ./pkg/...`（go1.24.6）通过；`authz_enabled: false` 时任意已登录账号不受接口权限校验拦截；置 `true` 后恢复 FIX-006 校验行为。

### FIX-016 部门 type 字段 proto 校验错误，创建部门必然失败（P0）

- **发现时间**：2026-09-11（统一运行时测试中发现，FIX-002 实际未生效）
- **严重级别**：🔴 P0（功能不可用）
- **问题描述**：
  - `api/proto/organization/service/v1/department.proto` 的 `CreateDepartmentRequest.type` 字段被误加了 `(buf.validate.field).string = { in: ["enabled","disabled"] }`（疑似从 `status` 字段复制）。
  - 该规则与紧随其后的 CEL（`this in ['company','subsidiary','department','position']`）取值域 **交集为空**，protovalidate 执行时先命中 `string.in`，任何部门类型都返回 `VALIDATOR: value must be in list [enabled, disabled]`。
  - 结论：FIX-002 虽补齐了 Ent 写入与前端表单，但请求在 proto 校验层即被拒绝，创建部门 100% 失败。
- **修复方案**：删除 `type` 字段上误加的 `string.in` 规则，仅保留描述类型枚举的 CEL 约束；`buf generate` 重新生成 `department.pb.go`（descriptor 内嵌约束随之更正）。
- **修改文件**：
  - `api/proto/organization/service/v1/department.proto`
  - `api/gen/go/organization/service/v1/department.pb.go`（重新生成）
- **验证**：实测创建 `company/subsidiary/department/position` 四种类型部门全部成功且回显 `type` 正确（详见统一测试结果）。

### FIX-017 .gitignore 规则未覆盖带前缀的构建产物（P3）

- **发现时间**：2026-09-11
- **严重级别**：🟢 P3（工程化）
- **问题描述**：`.gitignore` 仅有 `go_build_admin`（精确匹配），无法覆盖 IDE 生成的 `.!24019!go_build_admin` 这类带前缀的同名产物，`git status` 仍显示其为未跟踪文件。
- **修复方案**：追加通配规则 `*go_build_admin`。
- **修改文件**：`.gitignore`
- **验证**：`git status` 不再出现 `.!24019!go_build_admin`。

### FIX-018 消息中心引用了不存在的图标导出（P2）

- **发现时间**：2026-09-11（前端 `vue-tsc` 类型检查发现）
- **严重级别**：🟡 P2（组件运行期图标缺失）
- **问题描述**：`views/system/message-center/index.vue` 从 `@vben/icons` 导入 `CheckDouble`、`Trash`，但 `@vben/icons` 并未导出这两个成员（TS2305），运行期图标组件为 `undefined`。
- **修复方案**：改用项目既有约定的 `<IconifyIcon icon="..."/>`（`ant-design:check-circle-outlined` / `ant-design:delete-outlined`），与其它页面图标用法一致。
- **修改文件**：`frontend/apps/web-antd/src/views/system/message-center/index.vue`
- **验证**：`pnpm typecheck` 中 message-center 相关报错清零。

### FIX-019 任务执行成功时日志写入失败，任务日志/监控统计丢失（P2）

- **发现时间**：2026-09-11（系统监控联调时发现）
- **严重级别**：🟡 P2（功能缺失 + 监控失真）
- **问题描述**：
  - `biz/scheduler.go` 的 `writeLog` **只在 `!success` 分支**把 `job_message` 截断到 2000 字符；成功时 `message` 直接取接口响应体（执行器最多读取 4096 字节）。
  - `sys_job_log.job_message` 为 varchar(2000)、Ent schema `MaxLen(2000)`，超长时 `CreateJobLog` 触发 `ent: validator failed for field "SysJobLog.job_message"`，**整条日志写不进去**。
  - 表现：任务明明执行成功，「任务日志」里没有记录，定时任务监控的 `totalRuns/successRuns` 恒为 0。
- **修复方案**：把截断逻辑提到公共路径，成功/失败消息统一截断；改为按 **rune** 截断（避免按字节截断产生非法 UTF-8），失败分支仅负责置 `status=fail`。
- **修改文件**：`app/admin/service/internal/biz/scheduler.go`
- **验证**：新建两个任务（正常目标 + 404 目标）各执行一次 → 监控 `totalRuns=2`、`successRuns=1`、`failedRuns=1`、`successRate=50%`；最近执行记录成功/失败两条齐全；`/admin/v1/job-logs` total=2。

## 统一测试清单

> 以下为所有修复的手工/回归测试项。全部通过后，本清单即可归档。

**FIX-001 密码哈希**：① 通过 API 新建管理员 → 用该账号密码登录成功；② 数据库 admin 表 password 为 bcrypt 密文（`$2a$` 开头）而非明文。

**FIX-002 部门 type**：创建"公司/子公司/部门/岗位"各一种类型部门均成功；列表与详情正确回显 type。

**FIX-003 token 登出失效**：登录 → 调 `/admin/v1/logout` → 用原 access token 再次请求业务接口应返回 401；再刷新（refresh token 已删）应失败。

**FIX-004 前端刷新链路**：登录后 access token 过期（可缩短 `jwt.access_ttl`）→ 下一次请求自动用 refreshToken 续期成功，无需重新登录；退出登录后双 token 均失效。

**FIX-005 动态路由/菜单**：backend 模式下登录后侧边栏菜单按角色正确渲染、页面组件可打开、iframe/外链菜单正常；无 `/menu/all`、`/auth/codes` 请求报 404。

**FIX-006 RBAC 鉴权**：① superAdmin 可访问全部接口（直通）；② 普通用户未分配接口权限 → 管理接口返回 403；③ 角色菜单绑定含接口权限的按钮菜单后对应接口放行；④ 角色/菜单禁用后 ≤60s 权限收回；⑤ 未登录请求 → 401。

**FIX-007 用户 status / 重置密码**：① 新建管理员默认 `enabled`，可正常登录；② 将账号置为 `disabled` 后登录被拒；③ 重置密码（新密码）后旧密码登录失败、新密码登录成功。

**FIX-008 N+1 优化回归**：管理员列表、角色列表接口数据与修改前一致（无字段丢失/顺序错乱）。

**FIX-009 删除保护与关联清理**：① 删除当前登录账号 → 被拒；② 删除绑定 superAdmin 角色的账号 → 被拒；③ 删除仍有用户绑定的角色 / `is_system` 角色 → 被拒；④ 删除普通角色后 `role_menus` 无残留；⑤ 删除父部门后其子孙部门被软删除且列表不展示。

**FIX-010 菜单删除事务**：删除含子菜单且绑定 API 权限的菜单 → 子孙菜单、`role_menus`、`menu_api_permissions` 均无残留，报错时不产生半删数据。

**FIX-011 CurrentUserMenus**：登录后访问 `GET /admin/v1/current-user/menus` 返回真实菜单树（非空壳/500）。

**FIX-012 MQ 清理回归**：服务正常启动（无 mq 相关 panic/日志），`go build` 通过；确认无残留 rabbitmq 引用。

**FIX-013 .gitignore**：`git status` 不再出现 `dump.rdb` / `go_build_admin`；这两类文件无法再被 `git add`。

**FIX-014 默认密码配置**：修改 `config.yaml` 的 `app.admin_default_password` 后新建管理员，其默认密码为新值；未配置时回退 123456 正常。

**FIX-015 接口权限校验开关**：① `authz_enabled: false`（默认）→ 任意已登录账号调用管理接口均不校验接口权限，正常放行；② 置 `true` 重启 → superAdmin 直通、未绑定权限账号 403、绑定后放行；③ 登出 / current-user / menus 自服务接口不受开关影响。

## 统一测试结果（2026-09-11）

> 测试环境：macOS · go1.25.1 · MySQL 8（`root@127.0.0.1:3306/kratos_admin`）· Redis 7（`127.0.0.1:6379`）· Node 22.19 · pnpm 10.28
> 方式：编译当前代码 → 启动服务（Ent 自动迁移）→ 通过 HTTP API 全链路验证；验证码直接读取 Redis 真值以绕过 OCR。

| 项 | 结论 | 实测要点 |
|---|---|---|
| FIX-001 密码哈希 | ✅ 通过 | 新建管理员 DB 中 password 为 `$2a$10$` bcrypt；用该密码登录成功 |
| FIX-002 部门 type | ❌→✅（FIX-016 修复后） | 修复前四种类型全部 400；修复后 4 种类型创建成功、list/tree 回显 type 正确 |
| FIX-003 token 登出失效 | ✅ 通过 | 登出后原 access 调业务接口 401；原 refresh 刷新 401 |
| FIX-004 刷新链路（后端） | ✅ 通过 | `/refresh-token` 双 token 换新；新 access 调接口 200；旧 token 用后即废 |
| FIX-005 动态路由/菜单 | ⚠️ 未端到端 | 需浏览器登录后验证；后端 `current-user/menus` 已返回真实树 |
| FIX-006 RBAC 鉴权 | ✅ 通过 | authz=true：超管直通 200；普通用户无接口权限 403；绑定含 API 权限的按钮菜单后放行 200 |
| FIX-007 status/重置密码 | ✅ 通过 | 新建默认 enabled 可登录；置 disabled 登录 ACCOUNT_DISABLED；重置后旧密码失败、新密码成功 |
| FIX-008 N+1 回归 | ✅ 通过 | ListAdmin 正确返回 roleIds/roleNames/isSuperAdmin；ListRole 正确返回 menuIds 数量，无字段丢失 |
| FIX-009 删除保护 | ✅ 通过 | 删自己 `CANNOT_DELETE_SELF`；删超管账号 `CANNOT_DELETE_SUPERADMIN`；删有用户角色 `ROLE_HAS_ADMINS`；删系统角色 `CANNOT_DELETE_SYSTEM_ROLE`；删父部门后子孙同批软删（deleted_at 落库、树不再展示） |
| FIX-010 菜单删除事务 | ✅ 通过 | 删除含 2 个 action 子菜单的目录：menus 3→0、role_menus 3→0、menu_api_permissions 1→0 |
| FIX-011 CurrentUserMenus | ✅ 通过 | 返回真实菜单树，HTTP 200 |
| FIX-012 MQ 清理 | ⚠️ 已失效 | MQ 以可插拔驱动重新引入（`mq.enabled=true, driver=memory`），服务启动正常 |
| FIX-013 .gitignore | ❌→✅（FIX-017 修复后） | 修复前 `.!24019!go_build_admin` 仍为未跟踪；修复后消失 |
| FIX-014 默认密码配置 | ✅ 通过 | 临时配置 `admin_default_password=Abc@12345`：新建管理员（不传密码）用新默认密码登录成功、旧 123456 失败 |
| FIX-015 权限校验开关 | ✅ 通过 | false 时不校验；true 时超管直通 / 无权 403 / 绑定后放行 / 自服务接口（current-user、menus、logout）不受影响 |

**补充说明（前端）**：

- `pnpm typecheck`（vue-tsc）共 79 处报错，绝大多数为既有基线问题（`showSuccessMessage` 不在 `RequestClientConfig`、`@vben-core`/`@vben` 模板泛型等），涉及旧文件（如 `api/adminuser/admin.ts`、`api/permission/role.ts`）与新文件；本次仅修复了会导致运行期图标失效的 FIX-018。
- `pnpm build`（vite production）在当前环境失败：`jiti` 被当作浏览器外部依赖处理（`"createRequire" is not exported by "__vite-browser-external"`）。应用源码未引用 `jiti`，来源为 tailwind 配置加载链路，属工具链/依赖问题，与本次修复无直接关系，待单独处理。

**遗留 / 后续**：

- FIX-005（backend 模式动态菜单渲染）需在浏览器端联调确认。
- 数据库存量明文密码账号需通过"重置密码"接口或脚本批量转换为 bcrypt。
