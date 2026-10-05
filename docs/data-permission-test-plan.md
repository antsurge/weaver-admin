# 数据权限手动测试方案

> 本文档为「数据权限」功能的手动测试指导，覆盖规则管理、绑定链路、范围解析、边界场景、账号设计及验收标准。
> 适用版本：weaver-admin（数据权限规则化改造后，5 种范围 = 全部 / 自定义 / 本部门 / 本部门及以下 / 仅本人）。

---

## 一、当前数据权限的实现边界（先明确，避免测错方向）

当前实现是 **「配置 + 绑定 + 解析下发」** 模式：

| 环节 | 实现 | 位置 |
|---|---|---|
| 规则 CRUD（5 种范围） | 完整 | `biz/data_permission.go` + 「权限管理 → 数据权限」菜单 |
| 用户绑定规则 | 完整 | `admin_data_permission` 关联表，用户表单「数据权限」多选 |
| 角色绑定规则 | 完整 | `role_data_permission` 关联表，角色表单 |
| 生效范围计算 | 完整 | `resolveDataScope`：用户绑定 > 角色并集 > 默认仅本人；超管恒为 `ALL` |
| 下发前端 | 完整 | `current-user` 接口返回 `dataScope` 字段 |
| **列表查询时真正过滤数据** | **未落地** | `ListAdmin` 等内置查询未按 dataScope 过滤，前端也未消费 dataScope 裁剪列表 |

> ⚠️ 注意：当前版本能测的是「规则管理、绑定、范围解析、下发正确性」；「业务数据被数据权限裁剪」尚未在系统内置接口（如用户列表）上实现。如需验证数据裁剪效果，需通过「代码生成器」生成带 `department_id` / `created_by` 字段的业务表，并自行接入 `dataScope` 消费逻辑。

---

## 二、测试前置准备

### 1. 组织数据（部门树）

```
总公司 (root)
├── 华东分公司
│   ├── 上海部门
│   └── 杭州部门
└── 华南分公司
    ├── 深圳部门
    └── 广州部门
```

### 2. 数据权限规则（「权限管理 → 数据权限」创建）

| 规则 code | 名称 | scope_type | 自定义范围 |
|---|---|---|---|
| `ALL` | 全部数据 | 1 全部 | — |
| `DEPT` | 本部门数据 | 3 本部门 | — |
| `DEPT_AND_CHILD` | 本部门及以下 | 4 | — |
| `SELF` | 仅本人 | 5 | — |
| `CUSTOM_SH` | 自定义-上海及杭州 | 2 自定义 | deptIds=[上海部门, 杭州部门] |
| `CUSTOM_ROLE` | 自定义-按角色 | 2 自定义 | roleIds=[华东角色] |

### 3. 角色（「权限管理 → 角色管理」创建）

| 角色 | 说明 |
|---|---|
| 普通角色（华东） | 绑定菜单权限，可绑数据权限规则 |
| 普通角色（无规则） | 不绑任何数据权限规则（验证默认行为） |
| 超管角色 | 已有 `SuperAdmin`（isSystem / isSuperAdmin），编码配置在 `config.yaml` 的 `super_admin_code` |

---

## 三、测试账号设计（核心）

用以下 7 个账号覆盖全部 5 种范围 + 组合场景：

| # | 用户名 | 归属部门 | 绑定角色 | 用户级规则 | 期望 dataScope |
|---|---|---|---|---|---|
| A1 | `admin` | 根 | SuperAdmin | — | `ALL`（超管恒为全部） |
| A2 | `all_user` | 上海部门 | 普通角色+规则`ALL` | — | `ALL` |
| A3 | `dept_user` | 上海部门 | 普通角色+规则`DEPT` | — | `DEPT` |
| A4 | `child_user` | 华东分公司 | 普通角色+规则`DEPT_AND_CHILD` | — | `DEPT_AND_CHILD`（含上海/杭州） |
| A5 | `self_user` | 杭州部门 | 普通角色（无规则） | — | `SELF`（默认仅本人） |
| A6 | `custom_user` | 深圳部门 | 普通角色+规则`CUSTOM_SH` | — | `DEPT_AND_CHILD`（自定义按部门集合） |
| A7 | `override_user` | 广州部门 | 普通角色+规则`SELF` | **用户级规则 `ALL`** | `ALL`（**用户级 > 角色级**，验证覆盖优先级） |

另加 2 个边界账号：

| # | 用户名 | 场景 | 期望 |
|---|---|---|---|
| A8 | `norole_user` | 不绑定任何角色、任何规则 | `SELF`（兜底值） |
| A9 | `multirole_user` | 绑定「普通角色+DEPT」+「普通角色+ALL」 | `ALL`（**多角色取并集 = 权限最大者**） |

---

## 四、测试步骤（核心测试矩阵）

### 测试 1：规则管理 CRUD（用 admin 执行）

1. 创建上面 6 条规则 → 校验唯一 code 报错（重复 code 应提示「规则编码已存在」）
2. 编辑规则改 scope_type 为自定义并选部门集合 → 保存
3. 禁用规则（状态开关）→ 确认禁用后不被 `resolveDataScope` 采用
4. 删除规则（软删除）→ 已绑定用户/角色不再生效

### 测试 2：绑定链路（用 admin 执行）

1. 用户 A3 表单绑定规则 `DEPT` → 用户列表显示「本部门数据」标签
2. 角色绑定规则 → 用户 A4 通过角色继承获得规则
3. 用户 A7 同时设置「用户级 ALL」+「角色级 SELF」→ **用户级覆盖角色级**

### 测试 3：范围解析正确性（验证 current-user 接口）

对 A1~A9 分别登录，调用 `GET /admin/v1/current-user`，断言返回的 `dataScope`：

```
A1=ALL  A2=ALL  A3=DEPT  A4=DEPT_AND_CHILD  A5=SELF
A6=DEPT_AND_CHILD  A7=ALL  A8=SELF  A9=ALL
```

### 测试 4：禁用规则联动（用 admin 执行）

1. 将 A2 角色绑定的 `ALL` 规则禁用 → 重新登录 A2 → dataScope 降级为 `SELF`
2. 将 A7 用户级规则 `ALL` 删除 → dataScope 回落到角色级 `SELF`

### 测试 5：超管保护

1. 用 admin 尝试删除 `admin`（超管账号）→ 应报「超级管理员账号不允许删除」
2. 尝试删除自己 → 应报「不能删除当前登录账号」
3. 给任意用户绑定 SuperAdmin 角色 → 该用户 dataScope 恒为 `ALL`

### 测试 6（可选）：业务数据过滤验证

用「代码生成器」生成一张含部门/归属人字段的业务表，并在其列表查询中消费当前用户 `dataScope`（ALL 不过滤 / DEPT 按 `department_id` 过滤 / SELF 按 `created_by` 过滤），再按 A1~A7 登录验证数据裁剪。

---

## 五、不需要区分数据权限的菜单/数据

数据权限针对的是**有「归属部门/归属人」维度的业务数据**。以下属于**全局系统配置/共享数据**，一律**不做**数据权限过滤（且应授权给管理员可见）：

| 模块 | 原因 |
|---|---|
| 菜单管理 | 菜单树是权限配置本身，被过滤会造成权限管理死锁 |
| 角色管理 | 角色是授权配置，全量可见 |
| 数据权限规则管理 | 规则库自身，全量可见 |
| 字典管理 | 全局共享字典，与部门无关 |
| 参数设置 | 全局配置项 |
| 接口管理 | API 元数据，全量 |
| 租户管理 | 租户级数据，由租户隔离控制而非数据权限 |
| 消息通知 / 消息中心 | 管理端全局发布 |
| 定时任务 / 任务日志 | 系统运维数据 |
| 系统监控（服务/缓存/数据库/接口/任务） | 运维数据，全量 |
| 在线用户 / 操作日志 / 登录日志 | 安全审计，全量 |
| 低代码（表单构建器 / 代码生成器） | 开发工具，全量 |

**需要区分权限的数据**（当前内置模块中）：

- 用户列表（按归属部门裁剪）
- 部门管理 / 职务管理（按组织范围裁剪，一般配合部门树局部展开）
- 未来通过代码生成器建立的各类业务表（带 `department_id` / `created_by` 字段）

---

## 六、验收标准（一张表对照）

| 场景 | 期望结果 |
|---|---|
| 绑定超管角色 | dataScope=ALL，且不可被删除 |
| 用户级规则存在 | 优先于角色级规则 |
| 用户级规则为空 | 回退到角色规则并集（取权限最大者） |
| 无角色无规则 | dataScope=SELF（兜底） |
| 多规则/多角色 | 取最大范围（并集效果） |
| 规则被禁用/删除 | 立即失效，重新登录后范围回落 |
| 全部数据范围 | 不做任何过滤，看到全部 |
| 仅本人范围 | 只能看到自己创建/归属自己的数据 |

---

## 七、相关代码位置（排查/二次开发参考）

| 功能 | 位置 |
|---|---|
| 规则 CRUD（biz） | `app/admin/service/internal/biz/data_permission.go` |
| 规则 CRUD（data） | `app/admin/service/internal/data/data_permission.go` |
| 用户绑定规则 | `app/admin/service/internal/data/admin.go`（`admin_data_permission` 关联表） |
| 角色绑定规则 | `app/admin/service/internal/data/role.go`（`GetDataPermissionsByRoleIDs`） |
| 范围解析下发 | `app/admin/service/internal/biz/authentication.go`（`resolveDataScope`） |
| 范围常量/映射 | `app/admin/service/internal/biz/data_permission.go`（`DataScopeFromRule` / `DataScopeRank`） |
| 前端规则管理页 | `frontend/apps/web-antd/src/views/permission/data-permission/` |
| 前端用户绑定表单 | `frontend/apps/web-antd/src/views/adminuser/admin/modules/form/index.vue` |
