# MiniPMS 页面信息架构（V1）

> 依据：`mvp/features.md`、`mvp/schema.sql`  
> 状态：已确认  
> 锁定：建产品后留产品详情 · 需求列表强制选产品 · 迭代仅拉入 `type=story` 且 `status=active`

---

## 0. 总体结构

### 0.1 布局

| 区域 | 内容 |
|------|------|
| 未登录 | 仅登录页，无侧栏 |
| 已登录 | 顶栏 + 左侧菜单 + 主内容区 |
| 顶栏 | 系统名、当前用户、改密入口、登出 |
| 侧栏 | 按用户菜单树渲染（`dir` 折叠，`menu` 可点；`button` 不进侧栏） |

### 0.2 路由总表

| 路由 | 页面 | 菜单 code | 功能 ID |
|------|------|-----------|---------|
| `/login` | 登录 | — | F-AUTH-01 |
| `/dashboard` | 工作台 | `dashboard` | — |
| `/products` | 产品列表 | `product.list` | F-PROD-01 |
| `/products/new` | 新建产品 | `product.create` | F-PROD-02/06 |
| `/products/:id` | 产品详情 | `product.list` | F-PROD-03~05 |
| `/products/:id/edit` | 编辑产品 | `product.edit` | F-PROD-03 |
| `/stories` | 需求列表 | `story.list` | F-STORY-01 |
| `/stories/new` | 新建需求 | `story.create` | F-STORY-02 |
| `/stories/:id` | 需求详情 | `story.list` | F-STORY-03~09 |
| `/stories/:id/edit` | 编辑需求 | `story.edit` | F-STORY-03~05 |
| `/projects` | 项目列表 | `project.list` | F-PRJ-01 |
| `/projects/new` | 新建项目 | `project.create` | F-PRJ-02 |
| `/projects/:id` | 项目详情 | `project.list` | F-PRJ-03~05 |
| `/projects/:id/edit` | 编辑项目 | `project.edit` | F-PRJ-03/04 |
| `/sprints` | 迭代列表 | `sprint.list` | F-SPR-01 |
| `/sprints/new` | 新建迭代 | `sprint.create` | F-SPR-02 |
| `/sprints/:id` | 迭代详情（含关联需求） | `sprint.list` | F-SPR-03~06 |
| `/sprints/:id/edit` | 编辑迭代 | `sprint.edit` | F-SPR-03 |
| `/bugs` | 缺陷列表 | `bug.list` | F-BUG-01 |
| `/bugs/new` | 新建缺陷 | `bug.create` | F-BUG-02 |
| `/bugs/:id` | 缺陷详情 | `bug.list` | F-BUG-03~10 |
| `/bugs/:id/edit` | 编辑缺陷 | `bug.edit` | F-BUG-03 |
| `/system/users` | 用户列表 | `user.list` | F-USER-01~05 |
| `/system/roles` | 角色列表 | `role.list` | F-ROLE-01~03 |
| `/system/roles/:id/menus` | 分配菜单 | `role.assignMenu` | F-ROLE-03 |
| `/system/menus` | 菜单管理 | `menu.list` | F-MENU-01~04 |
| `/profile/password` | 修改密码 | —（登录即可） | F-AUTH-04 |

无独立「附件页」：附件嵌在需求/缺陷详情内。

### 0.3 交互约定

1. **无权限**：侧栏不显示；直链进入显示 403 页。  
2. **删除/危险操作**：二次确认对话框。  
3. **新建成功**：优先进详情页；建产品成功**留在产品详情**，Toast 提示「已自动创建项目 xxx1.0」，提供跳转该项目的链接（不自动跳转）。  
4. **人员选择**：下拉展示 `realname (account)`，存 `user_id`。  
5. **列表分页**：默认每页 20；支持关键词搜索处单独注明。  
6. **弹窗 vs 独立页**：分配角色、解决缺陷、关联需求用**抽屉/弹窗**；主实体新建/编辑用**独立页**。

---

## 1. 登录 `/login`

| 项 | 内容 |
|----|------|
| 字段 | account*、password* |
| 按钮 | 登录 |
| 成功 | 跳转 `/dashboard`（或 redirect 参数） |
| 失败 | 统一错误文案 |

---

## 2. 工作台 `/dashboard`

只读卡片（有对应 list 权限才显示块）：

| 卡片 | 内容 | 跳转 |
|------|------|------|
| 我的需求 | `assigned_to=我` 且 status≠closed，最近 5 条 | `/stories?assignedTo=me` |
| 我的缺陷 | 同上 | `/bugs?assignedTo=me` |
| 进行中迭代 | status=doing，最近 5 条 | `/sprints?status=doing` |
| 快捷入口 | 按权限：新建产品/需求/缺陷 | 对应 new 路由 |

---

## 3. 产品

### 3.1 列表 `/products`

**筛选：** status（全部/normal/closed）、关键词（name/code）

**列：** ID、名称、代号、状态、负责人(PO)、项目数、需求数、创建时间、操作

**行操作 / 顶栏按钮：**

| 按钮 | 权限 | 行为 |
|------|------|------|
| 新建 | `product.create` | → `/products/new` |
| 查看 | `product.list` | → 详情 |
| 编辑 | `product.edit` | → 编辑页 |
| 关闭/启用 | `product.edit` | normal↔closed |
| 删除 | `product.delete` | 确认后软删 |

### 3.2 新建 `/products/new`

| 字段 | 必填 | 说明 |
|------|:----:|------|
| name | ✓ | |
| code | | 唯一 |
| po | | 用户选择 |
| description | | 多行 |

提交成功：创建产品 + 自动项目 `{name}1.0` → 产品详情，Toast 含项目链接。

### 3.3 编辑 `/products/:id/edit`

同新建字段；不可改系统生成逻辑。关闭状态可在此改或详情操作。

### 3.4 详情 `/products/:id`

**头信息：** 名称、代号、状态、PO、描述、创建人/时间  

**Tab：**

| Tab | 内容 |
|-----|------|
| 概览 | 头信息 + 操作按钮（编辑/关闭/删除） |
| 项目 | 该产品下项目列表（精简列）→ 可跳项目详情；有 `project.create` 显示「新建项目」 |
| 需求 | 嵌入需求列表（product_id 固定） |
| 缺陷 | 嵌入缺陷列表（product_id 固定） |

---

## 4. 需求

### 4.1 列表 `/stories`

**筛选：** **产品必选**（无 `productId` 时展示空态：「请先选择产品」+ 产品下拉；从产品详情进入须带 `?productId=`）；type、status、指派人、关键词(title)

**列：** ID、标题、产品、类型(planning/story)、优先级、状态、估算、指派人、附件数、更新时间、操作

**按钮：** 新建 → `/stories/new`；行内查看/编辑/删除

**建议默认：** 从产品详情点「需求」进入时带 `?productId=`；全局进 `/stories` 必须先选产品后才加载列表。

### 4.2 新建 `/stories/new`

| 字段 | 必填 | 说明 |
|------|:----:|------|
| product_id | ✓ | 关闭中的产品不可选 |
| type | ✓ | 默认 `planning` |
| title | ✓ | |
| description | | |
| pri | | 默认 3 |
| estimate | | |
| assigned_to | | |
| 附件 | | 可先建后在详情上传；或创建成功后支持立即上传 |

### 4.3 编辑 `/stories/:id/edit`

可改：title、description、pri、estimate、assigned_to、type、status（按状态机可选值）

### 4.4 详情 `/stories/:id`

**头：** 标题、产品、类型、状态、pri、estimate、指派人  

**操作按钮：**

| 按钮 | 权限 | 说明 |
|------|------|------|
| 编辑 | `story.edit` | |
| 激活 | `story.edit` | draft→active |
| 关闭 / 重开 | `story.edit` | |
| 转为可交付 | `story.edit` | planning→story |
| 删除 | `story.delete` | |
| 上传附件 | `story.attach` | |

**区块：**
1. 描述  
2. 附件列表：文件名、大小、上传人、时间；预览(图)/下载；删除(`story.attach`)  
3. 关联信息：所在迭代列表（只读，来自 sprint_story）；相关缺陷入口  

---

## 5. 项目

### 5.1 列表 `/projects`

**筛选：** 产品、status、关键词  

**列：** ID、名称、代号、所属产品、状态、PM、开始/结束、迭代数、操作  

**按钮：** 新建；行内查看/编辑/删除  

### 5.2 新建 `/projects/new`

| 字段 | 必填 | 说明 |
|------|:----:|------|
| product_id | ✓ | 仅 normal 产品 |
| name | ✓ | |
| code | | |
| begin / end | | |
| pm | | |
| description | | |

> 自动创建的「xxx1.0」与手动新建并存。

### 5.3 编辑 `/projects/:id/edit`

可改 name/code/日期/pm/description/status；**不可改 product_id**。

### 5.4 详情 `/projects/:id`

**Tab：**
| Tab | 内容 |
|-----|------|
| 概览 | 基本信息 + 状态操作 |
| 迭代 | 本项目迭代列表；新建迭代带 `?projectId=` |
| 需求 | 本项目下所有已被某迭代拉入的需求（去重汇总，只读+跳转） |
| 缺陷 | `project_id` 筛选的缺陷 |

---

## 6. 迭代

### 6.1 列表 `/sprints`

**筛选：** 产品（可选，经项目间接）、项目、status  

**列：** ID、名称、项目、产品(冗余展示)、状态、开始/结束、需求数、操作  

### 6.2 新建 `/sprints/new`

| 字段 | 必填 | 说明 |
|------|:----:|------|
| project_id | ✓ | |
| name | ✓ | |
| begin / end | | 支持任意周期 |
| goal | | |

### 6.3 详情 `/sprints/:id`（核心页）

**头：** 名称、项目、状态、日期、目标  

**状态按钮：** 开始(doing) / 完成(done) / 关闭(closed) — 需 `sprint.edit`  

**Tab / 区：**

#### A. 迭代需求（`sprint.linkStory`）

- 已关联需求表：ID、标题、类型、状态、指派人、操作「移除」  
- **关联需求**按钮：仅 `status=doing` 时可用；打开弹窗  

**关联需求弹窗：**
- 数据源：同产品、**仅 `type=story` 且 `status=active`**、未在本迭代中的需求  
- 多选 + 确认写入 `sprint_story`  
- 非 doing 时按钮灰显并提示原因  
- `planning` 或不在 active 的需求不出现在候选列表（须先在需求侧转换/激活）  

#### B. 相关缺陷

- 列表筛选 `sprint_id=:id`；快捷「新建缺陷」带上 product/project/sprint 默认值  

---

## 7. 缺陷

### 7.1 列表 `/bugs`

**筛选：** 产品、项目、迭代、需求、status、severity、pri、指派人、关键词  

**列：** ID、标题、产品、严重程度、优先级、状态、指派人、关联需求、附件数、操作  

### 7.2 新建 `/bugs/new`

| 字段 | 必填 | 说明 |
|------|:----:|------|
| product_id | ✓ | |
| title | ✓ | |
| steps | | |
| severity / pri | | 默认 3 |
| project_id | | 须属该产品 |
| sprint_id | | 须属该项目 |
| story_id | | 须属该产品 |
| assigned_to | | |
| 附件 | | 可保存后上传 |

级联选择：产品 → 项目 → 迭代；需求按产品过滤。

### 7.3 详情 `/bugs/:id`

**操作：** 编辑、解决(弹窗填 resolution)、关闭、激活、删除(仅 active)、上传附件  

**区块：** 步骤描述、附件区、关联对象链接（产品/项目/迭代/需求）

**解决弹窗字段：** resolution*（枚举）

---

## 8. 系统管理

### 8.1 用户 `/system/users`

**筛选：** status、关键词(account/realname)  

**列：** ID、账号、姓名、邮箱、状态、角色(标签)、操作  

| 操作 | 权限 | 形态 |
|------|------|------|
| 新建 | `user.create` | 弹窗/页：account*、password*、realname*、email、角色多选 |
| 编辑 | `user.edit` | 不可改 account；可改姓名邮箱状态 |
| 停用/启用 | `user.disable` | |
| 分配角色 | `user.assignRole` | 多选角色弹窗 |

### 8.2 角色 `/system/roles`

**列：** 名称、code、builtin、备注、操作  

| 操作 | 权限 |
|------|------|
| 编辑名称/备注 | `role.edit`（code 只读） |
| 分配菜单 | `role.assignMenu` → `/system/roles/:id/menus` |

**分配菜单页：** 树形勾选 dir/menu/button；保存覆盖 `role_menu`；manager 保底校验。

### 8.3 菜单 `/system/menus`

树表：名称、code、type、path、sort、status  

支持新建子节点/编辑/删除（有子则禁删）。MVP 可以经理维护为主。

### 8.4 改密 `/profile/password`

字段：旧密码*、新密码*、确认*；成功后可保持登录或要求重登（实现二选一，建议保持登录）。

---

## 9. 关键跳转流

```text
登录 → 工作台
  ├─ 新建产品 → 产品详情（提示自动项目）→ 项目详情 → 新建迭代 → 迭代详情(doing) → 关联需求
  ├─ 产品详情.需求 → 新建需求(带 productId) → 需求详情(上传附件)
  └─ 迭代详情 → 新建缺陷(带 product/project/sprint) → 缺陷详情(上传截图)
```

面包屑建议：

- 产品 / 产品名 / 需求#id  
- 产品 / 产品名 / 项目名 / 迭代名  

---

## 10. 页面 ↔ 权限码速查

| 页面元素 | code |
|----------|------|
| 侧栏产品列表 | `product.list` |
| 按钮新建产品 | `product.create` |
| 按钮编辑/关闭产品 | `product.edit` |
| 按钮删除产品 | `product.delete` |
| 侧栏需求 | `story.list` |
| 新建/编辑/删需求 | `story.create/edit/delete` |
| 需求附件 | `story.attach` |
| 侧栏项目/迭代 | `project.list` / `sprint.list` |
| 项目 CRUD | `project.create/edit/delete` |
| 迭代 CRUD | `sprint.create/edit/delete` |
| 关联需求 | `sprint.linkStory` |
| 缺陷列表与处理 | `bug.list/create/edit/resolve/close/delete/attach` |
| 系统三角色页 | `user.*` / `role.*` / `menu.*` |

---

## 11. 已锁定决策

| # | 项 | 选择 |
|---|----|------|
| 1 | 建产品成功后 | **A** 留在产品详情 + Toast 链到自动项目 |
| 2 | 需求列表 | **A** 强制先选产品 |
| 3 | 关联需求候选 | **A** 仅 `type=story` 且 `status=active` |

配套：`mvp/api.md`（接口契约）。  
下一步：选定技术栈并搭建工程。
