# weaver-admin 功能梳理（按菜单管理划分）

> 本文档与 `docs/feature-summary.md`（按功能域划分）互补，改为**站在"菜单管理"页的视角**，以数据库 `menu` 表中的菜单树为骨架，逐节点挂载对应的前端页面、后端接口与核心功能。
>
> **菜单来源口径**：`app/admin/service/internal/data/data.go` 的幂等种子（系统管理/租户管理/低代码）、`docs/sql/update-menu-components.sql`（component 路径映射）、`docs/backend-menu-reference.md`（录入参考）、`frontend/apps/web-antd/src/router/modules/*`（前端路由）及 `src/locales/langs/zh-CN/*.json`。菜单树为**动态下发**：登录后由 `GET /admin/v1/current-user/menus` 返回，前端 `accessMode='backend'` 渲染。

---

## 菜单树总览

```
控制台 (Dashboard)
├── 分析页 (Analytics)
└── 工作台 (Workspace)

权限管理 (Permission)
├── 菜单管理 (Menu)
├── 角色管理 (Role)
└── 数据权限 (DataPermission)

用户管理 (Adminuser)
└── 用户列表 (Admin)

安全中心 (Security)         ◀ 已从系统管理独立（本轮迁移）
├── 在线用户 (Online)
├── 操作日志 (OperationLog)
├── 登录日志 (LoginLog)
└── 安全策略 (Settings)

系统管理 (System)
├── 接口管理 (ApiInterface)
├── 字典管理 (Dictionary)
├── 参数设置 (Config)
├── 租户管理 (Tenant)
├── 消息通知 (Notification)
├── 消息中心 (MessageCenter)
├── 定时任务 (Job)
├── 任务日志 (JobLog)
├── 系统监控 (Monitor)
│   ├── 服务监控 (Server)
│   ├── 缓存监控 (Cache)
│   ├── 数据库监控 (Database)
│   ├── 接口监控 (ApiMetric)
│   └── 定时任务监控 (JobMetric)
└── 低代码 (Lowcode)
    ├── 表单构建器 (FormBuilder)
    └── 代码生成器 (Codegen)

组织管理 (Organization)
├── 部门管理 (Department)
└── 职务管理 (Position)
```

> **迁移状态**：安全中心已完成代码迁移（seed + SQL + 前端路由 + i18n），见下方「安全中心」章节与 `docs/sql/update-menu-security.sql`；消息中心/运维中心/租户/低代码等目录拆分待后续迭代。

---

## 一、控制台（Dashboard）

> 目录，`type=catalog`，`path=/dashboard`，`weight=0`，前端 `views/dashboard/`

| 子菜单 | code | path | component | 说明 |
|---|---|---|---|---|
| 分析页 | Analytics | /dashboard/analytics | dashboard/analytics/index | 数据概览大屏 |
| 工作台 | Workspace | /dashboard/workspace | dashboard/workspace/index | 个人工作台 |

后端：无对应业务接口，纯展示页。

---

## 二、权限管理（Permission）

> 目录，`path=/permission`，前端 `views/permission/`

| 子菜单 | code | path | component | 说明 |
|---|---|---|---|---|
| **菜单管理** | Menu | /permission/menu | permission/menu/index | 菜单树 CRUD（本文档的骨架来源） |
| **角色管理** | Role | /permission/role | permission/role/index | 角色 CRUD + 菜单授权 + 数据范围 |
| **数据权限** | DataPermission | /system/data-permission | permission/data-permission/index | 5 种数据范围配置 |

### 菜单管理
- **前端**：`views/permission/menu/index.vue`（树形表格 + 拖拽排序）
- **后端**：`PermissionService`（MenuTree / CreateMenu / UpdateMenu / DeleteMenu / UpdateMenuStatus）
- **核心功能**：
  - 菜单类型 5 种：`catalog / menu / iframe / link / action`
  - 树形结构、行拖拽排序（weight）
  - 按钮菜单（type=action）可挂接接口权限（apiPermissions），打通按钮级 `v-access`
  - 删除走事务：同步清理 `role_menus` 与 `api_permissions`

### 角色管理
- **后端**：`PermissionService`（ListRolesByUser / BindMenusForRole / ListMenusByRole 等）
- **核心功能**：
  - 角色 CRUD、启禁用、`isSystem` / `isSuperAdmin` 标记
  - 菜单授权：`role/{id}/menus` 全量替换
  - 数据范围（dataScope）配置，与"数据权限"菜单配套

### 数据权限
- **前端**：`views/permission/data-permission/index.vue`
- **后端**：`DataPermissionRepo` + biz `data_permission`
- **核心功能**：5 种范围 = 全部 / 自定义 / 本部门 / 本部门及以下 / 仅本人，可绑定部门、角色集合

---

## 三、用户管理（Adminuser）

> 目录，`path=/adminuser`，`weight=10`，前端 `views/adminuser/`

| 子菜单 | code | path | component | 说明 |
|---|---|---|---|---|
| 用户列表 | Admin | /adminuser/admin | adminuser/admin/index | 后台管理员管理 |

- **后端**：`AdminService`（用户 CRUD + 角色绑定）+ biz `admin`
- **核心功能**：
  - 管理员 CRUD、状态启停（status）、重置密码、批量启停
  - 角色分配；单用户可以绑定多角色
  - 保护规则：不能删除自己 / 超管、登录名唯一性校验
  - bcrypt 密码哈希，默认密码可配置
  - 多租户下配合租户额度（max_users）校验

---

## 四、安全中心（Security）—— 本轮已迁移

> 顶级目录，`code=Security`，`path=/security`，`type=catalog`，`weight=6`。
> **迁移来源**：原挂在系统管理下的在线用户/安全策略 2 个子菜单，已由 `data.go ensureSecurityMenu()`（新环境自动 seed）+ `docs/sql/update-menu-security.sql`（存量库重挂）迁移至本目录。操作日志/登录日志已进一步拆分至"日志中心"（见下节）。

| 子菜单 | code | path | component | 说明 |
|---|---|---|---|---|
| **在线用户** | Online | /security/online | security/online/index | 多设备会话管理 + 强制下线 |
| **安全策略** | Settings | /security/settings | security/settings/index | 失败锁定/IP黑白名单/Token 有效期 |

### 在线用户
- **前端**：`views/security/online/index.vue`；**后端**：`OnlineService`（ListOnlineUsers / ForceLogout）
- **功能**：按用户聚合的多设备会话管理、**强制下线**（一键踢掉全部会话）

### 安全策略
- **前端**：`views/security/settings/index.vue`
- **功能**（保存即生效）：
  - 登录失败锁定（次数 / 窗口 / 锁定时长）
  - IP 白名单 / 黑名单（支持 CIDR）
  - Token 有效期策略

> **迁移清单**：`app/admin/service/internal/data/data.go`（seed 常量 + `ensureSecurityMenu`）、`docs/sql/update-menu-security.sql`、`frontend/.../router/modules/security.ts`（新）、`system.ts`（移除安全类路由）、`locales/page.json`（zh-CN/en-US 新增 `page.security.*`）。

---

## 四·五、日志中心（Logs）—— 本轮新增

> 顶级目录，`code=Log`，`path=/logs`，`type=catalog`，`weight=7`。
> **迁移来源**：原挂在安全中心下的操作日志/登录日志 2 个子菜单，已由 `data.go ensureLogMenu()`（新环境自动 seed）+ `docs/sql/update-menu-logs.sql`（存量库重挂）迁移至本目录。

| 子菜单 | code | path | component | 说明 |
|---|---|---|---|---|
| **操作日志** | OperationLog | /logs/operation-log | log/operation-log/index | 操作审计（脱敏），`pkg/middleware/oplog` |
| **登录日志** | LoginLog | /logs/login-log | log/login-log/index | 登录成功/失败记录 |

### 操作日志
- **前端**：`views/log/operation-log/index.vue`
- **功能**：操作人 / 模块 / IP / 耗时，参数**脱敏**，列表 + 详情 + 清空 + 批量删除
- **机制**：`pkg/middleware/oplog` 中间件自动记录

### 登录日志
- **前端**：`views/log/login-log/index.vue`
- **功能**：登录成功/失败记录 + 失败原因

> **迁移清单**：`app/admin/service/internal/data/data.go`（seed 常量 + `ensureLogMenu`）、`docs/sql/update-menu-logs.sql`（新）、`docs/sql/update-menu-components.sql`（component `security/log/*` → `log/*`）、`frontend/.../router/modules/logs.ts`（新）、`security.ts`（移除日志路由）、`locales/page.json`（zh-CN/en-US 新增 `page.log.title`）、`views/security/log/` → `views/log/`。

---

## 五、系统管理（System）

> 目录，`path=/system`，`weight=5`，种子创建于 `data.go ensureDefaultMenus()`；**安全类菜单已迁出至安全中心**，余下为配置/消息/运维/低代码类

### 1. 接口管理（ApiInterface）

| item | 值 |
|---|---|
| path / component | /api-interface / system/api-interface/index |

- **功能**：上传 `openapi.yaml` 解析入库（service / tag / method / path），是角色/按钮授权接口权限的数据来源，形成"菜单 → 按钮 → 接口"的闭环
- **后端**：`ApiInterfaceRepo`（`internal/data/openapi_scanner`）

### 2. 字典管理（Dictionary）

| item | 值 |
|---|---|
| path / component | /system/dictionary（路由） / system/dictionary/dict-type/index |

- **功能**：字典**类型** + 字典**数据**（label / value / 权重 / 扩展 JSON）二级管理，业务页按编码消费
- **后端**：`DictTypeRepo` + `DictDataRepo`

### 3. 参数设置（Config）

| item | 值 |
|---|---|
| path / component | /system/config / system/config/index |

- **功能**：config key-value 参数配置（内置/外置标记），提供按键取值接口供业务使用

### 4. 租户管理（Tenant）

| item | 值 |
|---|---|
| path / component | /system/tenant / tenant/index |

- **功能**：多租户 CRUD
  - 租户编码唯一性、到期时间、最大用户/角色数
  - `single`（默认）/ `multi` 两种模式
  - 数据强隔离：`tenantguard` 中间件 + Ent 查询/变更拦截器自动注入 `tenant_id`；跨租户只读可配
- **前端**：`views/tenant/`（index.vue + modules/form + data.ts）

### 5. 消息通知（Notification）— 管理端

| item | 值 |
|---|---|
| path / component | /notification / message/notification/index |

- **功能**：
  - 消息发布（投放范围：全体 / 指定角色 / 指定用户）
  - 状态机：`draft → published → revoked`，支持撤回
  - SSE 实时推送：`GET /admin/v1/notifications/stream`
  - MQ 可插拔驱动：memory / redis / rabbitmq / kafka

### 6. 消息中心（MessageCenter）— 用户收件箱

| item | 值 |
|---|---|
| path / component | /system/message-center / message/message-center/index |

- **功能**：收件箱（未读/全部已读）、软删除进回收站、未读数统计
- **前端**：`views/message/message-center/`

### 7. 定时任务（Job）

| item | 值 |
|---|---|
| path / component | /system/job / ops/job/index |

- **功能**：HTTP 调用型任务（URL + Cron 表达式）
  - 错失策略：`immediately / once / ignore`
  - 并发控制、**立即执行**
  - 任务执行日志（关联下方"任务日志"）

### 8. 任务日志（JobLog）

| item | 值 |
|---|---|
| path / component | /system/job-log / ops/job-log/index |

- **功能**：每次任务的执行记录、成功/失败、耗时、返回值

### 9. 系统监控（Monitor）—— 二级目录

| 子菜单 | code | path | component | 监控内容 |
|---|---|---|---|---|
| 服务监控 | Server | /system/monitor/server | ops/monitor/server/index | Go runtime：CPU/内存/堆/GC/协程/运行时长 |
| 缓存监控 | Cache | /system/monitor/cache | ops/monitor/cache/index | Redis：版本/模式/内存/连接/命中率/OPS + 键浏览（TTL/类型） |
| 数据库监控 | Database | /system/monitor/database | ops/monitor/database/index | 连接池状态 + 各表行数/数据大小 |
| 接口监控 | ApiMetric | /system/monitor/api | ops/monitor/api/index | 各接口请求量/失败率/平均耗时（`pkg/middleware/apimetrics`） |
| 定时任务监控 | JobMetric | /system/monitor/job | ops/monitor/job/index | 任务数/执行次数/成功率/最近执行 |

### 10. 低代码（Lowcode）—— 二级目录

| 子菜单 | code | path | component | 说明 |
|---|---|---|---|---|
| 表单构建器 | FormBuilder | /system/lowcode/form-builder | system/form-builder/index | JSON Schema 表单定义 + 可视设计器 + 运行时渲染 + 提交记录 |
| 代码生成器 | Codegen | /system/lowcode/codegen | system/codegen/index | 扫描 information_schema → 配置模板 → 预览 → 下载 ZIP（Go 代码） |

---

## 六、组织管理（Organization）

> 目录，`path=/organization`，`weight=8`，前端 `views/organization/`

| 子菜单 | code | path | component | 说明 |
|---|---|---|---|---|
| 部门管理 | Department | /organization/department | organization/department/index | 公司/子公司/部门/岗位四类组织多级树，软删除子树 |
| 职务管理 | Position | /organization/position | organization/position/index | 职务配置，名称/编码唯一性校验 |

---

## 六、菜单编码速查表

| 顶级菜单 | code | path |
|---|---|---|
| 控制台 | Dashboard | /dashboard |
| 权限管理 | Permission | /permission |
| 用户管理 | Adminuser | /adminuser |
| 安全中心 | Security | /security |
| 日志中心 | Log | /logs |
| 系统管理 | System | /system |
| 组织管理 | Organization | /organization |

| 子菜单 | code | path / component |
|---|---|---|
| 菜单管理 | Menu | /permission/menu |
| 角色管理 | Role | /permission/role |
| 数据权限 | DataPermission | /system/data-permission → permission/data-permission/index |
| 用户列表 | Admin | /adminuser/admin |
| 接口管理 | ApiInterface | /api-interface |
| 字典管理 | Dictionary | /system/dictionary → system/dictionary/dict-type/index |
| 参数设置 | Config | /system/config |
| 租户管理 | Tenant | /system/tenant → tenant/index |
| 消息通知 | Notification | /notification → message/notification/index |
| 消息中心 | MessageCenter | /system/message-center → message/message-center/index |
| 操作日志 | OperationLog | /logs/operation-log → log/operation-log/index |
| 登录日志 | LoginLog | /logs/login-log → log/login-log/index |
| 在线用户 | Online | /system/online → security/online/index |
| 安全策略 | Security | /system/security → security/settings/index |
| 定时任务 | Job | /system/job → ops/job/index |
| 任务日志 | JobLog | /system/job-log → ops/job-log/index |
| 服务监控 | Server | /system/monitor/server → ops/monitor/server/index |
| 缓存监控 | Cache | /system/monitor/cache → ops/monitor/cache/index |
| 数据库监控 | Database | /system/monitor/database → ops/monitor/database/index |
| 接口监控 | ApiMetric | /system/monitor/api → ops/monitor/api/index |
| 任务监控 | JobMetric | /system/monitor/job → ops/monitor/job/index |
| 表单构建器 | FormBuilder | /system/lowcode/form-builder → system/form-builder/index |
| 代码生成器 | Codegen | /system/lowcode/codegen → system/codegen/index |
| 部门管理 | Department | /organization/department |
| 职务管理 | Position | /organization/position |

---

## 七、说明与口径

1. **数据权威性**：菜单实际以 `menu` 表为准，本文档依据 seed 代码（`data.go`）+ 参考文档重建；数据库中若手工增删过菜单，以此文档为参照核对。
2. **视图 vs 菜单**：前端 `views/` 已按业务域重组（security / log / message / ops / tenant / permission / lowcode），但菜单树仍挂在 `系统管理` 等目录下，二者通过 `component` 字段映射，不一一对应。
3. **权限链路**：菜单管理页（`type=button`/`action` 菜单）→ 接口权限（ApiPermission）→ 前端 `v-access` 指令，全链路由后端 `authz` 中间件校验（总开关 `authz_enabled`）。
4. **前端静态路由**：`src/router/modules/*` 仅为开发期兜底路由；生产侧菜单一律走后端动态下发。