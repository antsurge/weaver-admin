# weaver-admin 功能梳理

> 项目定位：基于 **Kratos（Go 微服务） + Ent（ORM）+ Wire（DI）+ Vue3 / vben-admin（前端）** 的中后台管理系统，在 RuoYi 体系基础上增强多租户、低代码、消息推送等能力。
>
> 仓库结构：后端单体应用 `app/admin/service`，公共能力沉淀在 `pkg/`，API 定义集中在 `api/`（proto），前端为 `frontend/`（pnpm monorepo）。

---

## 一、认证与访问控制

| 功能 | 说明 |
|---|---|
| **登录认证** | 图形验证码 + 账号密码 + 可选租户编码登录；JWT 双令牌（access / refresh），支持自动刷新续期、登出作废 |
| **RBAC 权限** | 用户 → 角色 → 菜单 → 按钮/接口权限 四层模型；后端按角色返回菜单树，`GET /current-user` 返回 `menuTree + roleCodes` 驱动前端动态路由 |
| **接口鉴权** | `pkg/middleware/authz` 按 `HTTP Method + 路径模板` 匹配 ApiPermission 进行校验；superAdmin 直通；有全局开关 `authz_enabled`（默认关闭，避免权限数据未配齐时阻塞联调） |
| **安全策略** | 登录失败锁定（次数 / 窗口 / 锁定时长）、IP 白/黑名单（支持 CIDR）、Token 有效期策略，保存即生效 |
| **在线用户** | 按用户聚合的多设备会话管理、强制下线（踢掉全部会话） |

## 二、系统管理

| 功能 | 说明 |
|---|---|
| **用户管理** | 后台管理员 CRUD、启禁用（status）、重置密码、批量启停、角色分配；保护规则（不能删自己 / 超管、登录名唯一性校验）、bcrypt 密码哈希、默认密码可配置 |
| **角色管理** | CRUD、启禁用、`is_system` / `isSuperAdmin` 标记、菜单授权（全量替换 `role/{id}/menus`）、数据范围（dataScope） |
| **菜单管理** | 树形结构，五种类型：`catalog / menu / iframe / link / action`；行拖拽排序；按钮菜单挂接接口权限（apiPermissions）；删除走事务（清理 role_menus + api_permissions） |
| **部门管理** | 公司 / 子公司 / 部门 / 岗位四类组织的多级树；软删除子树 |
| **岗位管理** | 职务配置，名称 / 编码唯一性校验 |
| **字典管理** | 字典类型 + 字典数据（label / value / 权重 / 扩展 JSON），支持按编码供业务页消费 |
| **参数配置** | config key-value，内置 / 外置标记，提供按键取值接口 |
| **接口管理** | 上传 `openapi.yaml` 解析入库（service / tag / method / path），支撑按钮级接口权限选择的闭环 |

## 三、组织与多租户

| 功能 | 说明 |
|---|---|
| **多租户** | 租户 CRUD（编码唯一、到期时间、最大用户/角色数、联系人）；`single`（默认）/ `multi` 两种模式；数据强隔离（`tenantguard` 中间件 + Ent 查询/变更拦截器自动注入 `tenant_id`），跨租户只读可配 |
| **数据权限** | 5 种数据范围：全部 / 自定义 / 本部门 / 本部门及以下 / 仅本人，可绑定部门、角色集合 |

## 四、系统运维监控（5 类）

| 监控 | 说明 |
|---|---|
| **服务监控** | Go runtime：CPU / 内存 / 堆 / GC / 协程 / 运行时长 |
| **缓存监控** | Redis：版本 / 模式 / 内存 / 连接 / 命中率 / OPS + 缓存键浏览（TTL / 类型） |
| **数据库监控** | 连接池状态 + 各表行数 / 数据大小统计 |
| **接口监控** | 各接口请求量 / 失败率 / 平均耗时（`pkg/middleware/apimetrics` 进程内统计） |
| **定时任务监控** | 任务数 / 执行次数 / 成功率 / 最近执行记录 |

## 五、日志审计

| 功能 | 说明 |
|---|---|
| **操作日志** | 操作人 / 模块 / IP / 耗时，参数脱敏，详情 / 清空 / 批量删除（`pkg/middleware/oplog` 自动记录） |
| **登录日志** | 成功 / 失败记录 + 失败原因 |

## 六、自动化与低代码

| 功能 | 说明 |
|---|---|
| **代码生成器** | 扫描 information_schema 业务表 → 读取字段元数据 → 生成配置（组件类型 / 列表 / 表单 / 必填 / 主键 / 查询条件）→ 预渲染预览 → 下载 ZIP（Go 后端代码） |
| **动态表单构建器** | JSON Schema 表单定义 CRUD + 可视化设计器 + 运行时渲染 + 提交数据收集 / 分页查询 |
| **定时任务** | HTTP 调用型任务（URL + Cron），错失策略 `immediately / once / ignore`、并发控制、立即执行、任务执行日志 |

## 七、消息与文件

| 功能 | 说明 |
|---|---|
| **消息通知** | 管理端发布（投放范围：全体 / 角色 / 用户）、撤回、状态机 draft / published / revoked；用户收件箱（已读 / 全部已读 / 软删进回收站 / 未读数）；**SSE 实时推送**（`/admin/v1/notifications/stream`）；MQ 可插拔驱动（memory / redis / rabbitmq / kafka） |
| **文件存储** | S3 兼容对象存储（阿里云 OSS / 腾讯云 COS / 七牛云 / MinIO）：上传、删除、私有桶预签名临时地址 |

## 八、技术架构分层

```
frontend/  apps/web-antd   Vue3 + Ant Design Vue + Vxe-Table（pnpm monorepo）
           ├─ 路由：accessMode='backend'，动态路由由后端 menuTree 驱动
           └─ 按钮权限：v-access authCode 指令 / hasAccessByCodes

api/       proto/ 51 个文件（admin/service/v1 的 18 个 i_*.proto 是唯一 HTTP facade 层，
                           统一前缀 /admin/v1/*；各领域 proto 定义数据模型）

app/admin/service/
           ├─ internal/conf     配置（conf.proto → config.yaml）
           ├─ internal/server   HTTP(8888) / gRPC(9999) / 后台任务(job/log/mq)
           ├─ internal/service  19 个 gRPC/HTTP 服务实现
           ├─ internal/biz      31 个业务模块（含 authz/scheduler/monitor/tenant_quota...）
           ├─ internal/data     Ent 仓储 + 27 个 schema 模型 + 多租户拦截器
           └─ internal/handler  自定义 HTTP handler（验证码/上传/SSE 等）

pkg/       公共包：middleware(auth/authz/ipfilter/bufvalidate/oplog/apimetrics/tenantguard)、
           storage、tenant、metrics、errors、enthelper、utils(crypto/excel/copierx/auth/uuid/message)
```

## 九、项目现状（来自 docs/bugfix-log.md）

- **已完成**：FIX-001 ~ FIX-019 全部修复并通过本机 MySQL + Redis 运行时测试（密码哈希、部门 type、token 登出失效、刷新链路、RBAC 鉴权、删除保护、N+1 优化、事务删除等），详见 `docs/bugfix-log.md`。
- **可插拔 MQ**：`config.yaml` 中 `mq.enabled=true, driver=memory`（单副本），多副本部署改 `redis` 保证跨实例推送。
- **待办遗留**：
  - 前端 FIX-005 动态菜单需浏览器端到端确认；
  - 前端 `vue-tsc` 有 79 处既有类型基线报错（非本次改动引入）；
  - `vite build` 因 tailwind / jiti 工具链问题暂不可用。
