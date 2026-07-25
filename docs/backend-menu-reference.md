# 后端菜单数据录入参考指南

> 基于 `frontend/apps/web-antd/src/router/routes/modules/` 下的前端路由定义整理
>
> **使用说明**: 在后端菜单管理系统中录入菜单时，请参照此表设置各字段值

---

## 📋 字段说明

| 后端字段 | 说明 | 示例 |
|---------|------|------|
| `name` | 菜单名称（显示在侧边栏） | 权限管理 |
| `code` | 菜单编码（用于路由 name） | Permission |
| `type` | 菜单类型 | `menu_dir` 或 `menu` |
| `path` | 路由路径 | /permission |
| `icon` | 图标（支持 iconify 格式） | lucide:layout-dashboard |
| `component` | 组件路径（相对于 views/ 目录） | permission/menu/index |
| `weight` | 排序权重（数字越小越靠前） | 1 |
| `parentID` | 父级菜单 ID | 父级菜单的 id |

### type 类型说明

| 值 | 说明 | component 设置 |
|----|------|----------------|
| `menu_dir` | 目录（分组，不对应页面） | 留空或 `BasicLayout` |
| `menu` | 菜单（对应一个页面） | 填写组件相对路径 |
| `button` | 按钮（不显示在菜单中） | 留空 |

---

## 🎯 完整菜单结构参考

### 1. 权限管理 (Permission)

```
权限管理 (menu_dir)
├── 菜单管理 (menu)
└── 角色管理 (menu)
```

#### 父级 - 权限管理

| 字段 | 值 |
|-----|---|
| name | 权限管理 |
| code | Permission |
| type | **menu_dir** |
| path | /permission |
| icon | lucide:layout-dashboard |
| component | *(留空)* |
| weight | 1 |

#### 子级 - 菜单管理

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 菜单管理 |
| code | Menu |
| type | **menu** |
| path | /permission/menu |
| icon | lucide:area-chart |
| component | **permission/menu/index** |
| weight | 1 |

#### 子级 - 角色管理

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 角色管理 |
| code | Role |
| type | **menu** |
| path | /permission/role |
| icon | lucide:area-chart |
| component | **permission/role/index** |
| weight | 2 |

---

### 2. 用户管理 (Adminuser)

```
用户管理 (menu_dir)
└── 用户列表 (menu)
```

#### 父级 - 用户管理

| 字段 | 值 |
|-----|---|
| name | 用户管理 |
| code | Adminuser |
| type | **menu_dir** |
| path | /adminuser |
| icon | lucide:layout-dashboard |
| component | *(留空)* |
| weight | 10 |

#### 子级 - 用户列表

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 用户列表 |
| code | Admin |
| type | **menu** |
| path | /adminuser/admin |
| icon | lucide:area-chart |
| component | **adminuser/admin/index** |
| weight | 1 |

---

### 3. 控制台/仪表盘 (Dashboard)

```
控制台 (menu_dir)
├── 分析页 (menu)
└── 工作台 (menu)
```

#### 父级 - 控制台

| 字段 | 值 |
|-----|---|
| name | 控制台 |
| code | Dashboard |
| type | **menu_dir** |
| path | /dashboard |
| icon | lucide:layout-dashboard |
| component | *(留空)* |
| weight | 0 |

#### 子级 - 分析页

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 分析页 |
| code | Analytics |
| type | **menu** |
| path | /dashboard/analytics |
| icon | lucide:area-chart |
| component | **dashboard/analytics/index** |
| weight | 1 |

#### 子级 - 工作台

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 工作台 |
| code | Workspace |
| type | **menu** |
| path | /dashboard/workspace |
| icon | carbon:workspace |
| component | **dashboard/workspace/index** |
| weight | 2 |

---

### 4. 系统管理 (System)

```
系统管理 (menu_dir)
└── 字典管理 (menu)
```

#### 父级 - 系统管理

| 字段 | 值 |
|-----|---|
| name | 系统管理 |
| code | System |
| type | **menu_dir** |
| path | /system |
| icon | lucide:layout-dashboard |
| component | *(留空)* |
| weight | 5 |

#### 子级 - 字典管理

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 字典管理 |
| code | Dictionary |
| type | **menu** |
| path | /system/dictionary |
| icon | lucide:area-chart |
| component | **system/dictionary/dict-type/index** |
| weight | 1 |

---

### 5. 组织管理 (Organization)

```
组织管理 (menu_dir)
├── 部门管理 (menu)
└── 岗位管理 (menu)
```

#### 父级 - 组织管理

| 字段 | 值 |
|-----|---|
| name | 组织管理 |
| code | Organization |
| type | **menu_dir** |
| path | /organization |
| icon | lucide:layout-dashboard |
| component | *(留空)* |
| weight | 8 |

#### 子级 - 部门管理

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 部门管理 |
| code | Department |
| type | **menu** |
| path | /organization/department |
| icon | lucide:area-chart |
| component | **organization/department/index** |
| weight | 1 |

#### 子级 - 岗位管理

| 字段 | 值 |
|-----|---|
| parentID | *父级的 id* |
| name | 岗位管理 |
| code | Position |
| type | **menu** |
| path | /organization/position |
| icon | lucide:area-chart |
| component | **organization/position/index** |
| weight | 2 |

---

## 🔍 快速查找表

| 菜单名称 | code | 路径 | component 路径 |
|---------|------|------|---------------|
| **权限管理** | Permission | /permission | *(目录)* |
| ├── 菜单管理 | Menu | /permission/menu | permission/menu/index |
| └── 角色管理 | Role | /permission/role | permission/role/index |
| **用户管理** | Adminuser | /adminuser | *(目录)* |
| └── 用户列表 | Admin | /adminuser/admin | adminuser/admin/index |
| **控制台** | Dashboard | /dashboard | *(目录)* |
| ├── 分析页 | Analytics | /dashboard/analytics | dashboard/analytics/index |
| └── 工作台 | Workspace | /dashboard/workspace | dashboard/workspace/index |
| **系统管理** | System | /system | *(目录)* |
| └── 字典管理 | Dictionary | /system/dictionary | system/dictionary/dict-type/index |
| **组织管理** | Organization | /organization | *(目录)* |
| ├── 部门管理 | Department | /organization/department | organization/department/index |
| └── 岗位管理 | Position | /organization/position | organization/position/index |

---

## ⚠️ 注意事项

### 1. Component 路径格式
- ✅ 正确：`permission/menu/index`
- ❌ 错误：`/permission/menu/index` （不要前导斜杠）
- ❌ 错误：`/menu` （路径太短，无法匹配到正确组件）

### 2. 图标格式
- 支持 [Iconify](https://iconify.design/) 所有图标集
- 常用图标集：
  - `lucide:` - Lucide Icons（推荐）
  - `ant-design:` - Ant Design Icons
  - `mdi:` - Material Design Icons
  - `carbon:` - Carbon Icons
- 示例：
  - `lucide:layout-dashboard`
  - `ant-design:setting-outlined`
  - `mdi:account-group`

### 3. 排序规则
- `weight` 数字越小越靠前
- 建议使用负数表示靠前的菜单
- 同级菜单按 weight 升序排列

### 4. Path 路径规范
- 父级路径：`/模块名`（如 `/permission`）
- 子级路径：`/父级路径/子菜单名`（如 `/permission/menu`）
- 路径必须唯一，不能重复

---

## 📝 录入步骤建议

1. **先创建父级菜单**（type=menu_dir）
   - 记录返回的 id
2. **再创建子级菜单**（type=menu）
   - parentID 填入父级的 id
3. **测试验证**
   - 清除浏览器缓存
   - 重新登录
   - 检查侧边栏是否正确显示
   - 点击菜单是否能加载页面

---

## 🐛 常见问题

### Q: 菜单显示了但点击后页面空白？
**A:** 检查 component 路径是否正确，确保与 `views/` 目录下的文件路径匹配。

### Q: 图标没有显示？
**A:** 检查图标名称是否正确，可以在 [Iconify](https://iconify.design/) 网站搜索确认。

### Q: 菜单顺序不对？
**A:** 调整 weight 值，数字越小越靠前。

### Q: 子菜单没有显示？
**A:** 检查 parentID 是否正确填入了父级菜单的 id。
