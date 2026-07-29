# MiniPMS 接口契约（V1）

> 依据：`schema.sql` / `features.md` / `pages.md`  
> 状态：已确认范围，待实现时按此对接  
> 风格：REST + JSON；与编程语言无关  
> 鉴权：登录后携带会话（Cookie Session **或** `Authorization: Bearer <token>`，实现二选一，本文用 `Auth` 统称）

---

## 0. 约定

### 0.1 Base URL

```
/api/v1
```

### 0.2 统一响应

**成功**

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

**失败**

```json
{
  "code": 40001,
  "message": "人类可读错误",
  "data": null
}
```

| HTTP | code 段 | 含义 |
|------|---------|------|
| 200 | 0 | 成功（业务失败也可用 200+非0，见下表；**推荐业务错误用 4xx + 非0 code**） |
| 401 | 401xx | 未登录 |
| 403 | 403xx | 无权限 |
| 404 | 404xx | 不存在 |
| 422 | 422xx | 参数/状态机校验失败 |
| 500 | 500xx | 服务器错误 |

常用业务 code：

| code | 含义 |
|------|------|
| 0 | 成功 |
| 40100 | 未登录 |
| 40300 | 无权限（缺 menu.code） |
| 40400 | 资源不存在 |
| 42200 | 通用校验失败 |
| 42201 | 状态不允许 |
| 42202 | 产品已关闭 |
| 42203 | 迭代非 doing，不可关联需求 |
| 42204 | 需求类型/状态不可拉入迭代 |
| 42205 | 存在关联，禁止删除 |
| 42206 | 附件类型或大小不合法 |
| 42207 | 仅 active 缺陷可删 |
| 42208 | 可交付需求锁定 / 备注不可改 |

### 0.3 列表查询通用

| 参数 | 说明 |
|------|------|
| page | 默认 1 |
| pageSize | 默认 20，最大 100 |
| keyword | 可选，各资源定义匹配字段 |

列表 `data`：

```json
{
  "list": [],
  "page": 1,
  "pageSize": 20,
  "total": 100
}
```

### 0.4 人员展示

接口返回指派人等字段时，推荐嵌套：

```json
{
  "assignedTo": 3,
  "assignee": { "id": 3, "account": "dev1", "realname": "张三" }
}
```

写入只传 id。

### 0.5 权限

除登录、改密外，每个写操作/敏感读需校验对应 `menu.code`。缺权限 → 40300。  
前端用 `GET /auth/me` 返回的 `menus` 控制显隐；后端必须再验一次。

### 0.6 软删

默认列表/详情不返回 `deleted=1`。删除均为软删。

---

## 1. 认证 Auth

### POST `/auth/login`

匿名。F-AUTH-01

```json
// req
{ "account": "admin", "password": "xxx" }

// data
{
  "token": "<若用 Bearer>",
  "user": { "id": 1, "account": "admin", "realname": "系统管理员" }
}
```

### POST `/auth/logout`

需 Auth。F-AUTH-02 → `data: null`

### GET `/auth/me`

需 Auth。F-AUTH-03

```json
{
  "user": { "id": 1, "account": "admin", "realname": "...", "email": "..." },
  "roles": [{ "id": 4, "code": "manager", "name": "经理" }],
  "menus": [
    {
      "id": 2, "code": "product", "name": "产品", "type": "dir", "path": null, "icon": "product",
      "children": [
        { "id": 10, "code": "product.list", "type": "menu", "path": "/products", "children": [
          { "id": 100, "code": "product.create", "type": "button", "path": null }
        ]}
      ]
    }
  ]
}
```

`menus` 为当前用户权限并集后的树（含 button）。

### PUT `/auth/password`

需 Auth。F-AUTH-04

```json
{ "oldPassword": "...", "newPassword": "..." }
```

---

## 2. 用户 Users（`user.*`）

### GET `/users`

`user.list`  
Query: `status?`, `keyword?`(account/realname), page

### POST `/users`

`user.create`  
Body: `account*`, `password*`, `realname*`, `email?`, `roleIds?: number[]`

### GET `/users/:id`

`user.list`

### PUT `/users/:id`

`user.edit`  
Body: `realname?`, `email?`, `status?`（不可改 account）

### POST `/users/:id/disable` / `POST `/users/:id/enable``

`user.disable`

### PUT `/users/:id/roles`

`user.assignRole`  
Body: `{ "roleIds": [1,2] }` 覆盖式

---

## 3. 角色 / 菜单

### GET `/roles`

`role.list` → 含 `builtin`

### PUT `/roles/:id`

`role.edit`  
Body: `name?`, `remark?`（不可改 code；builtin 可改名）

### GET `/roles/:id/menus`

`role.assignMenu` → 返回该角色已选 `menuIds: number[]`

### PUT `/roles/:id/menus`

`role.assignMenu`  
Body: `{ "menuIds": [1,2,10,...] }`  
约束：manager 角色不可导致无法访问系统（F-SYS-08）

### GET `/menus/tree`

`menu.list` → 全量菜单树（管理用）

### POST `/menus`

`menu.create`  
Body: `parentId?`, `code*`, `name*`, `type*`, `path?`, `icon?`, `sort?`, `status?`

### PUT `/menus/:id`

`menu.edit`

### DELETE `/menus/:id`

`menu.delete`；有子节点 → 42205

---

## 4. 产品 Products

### GET `/products`

`product.list`  
Query: `status?`, `keyword?`(name/code)

`list[]` 建议含：`id,name,code,status,po,poUser,projectCount,storyCount,createdAt`

### POST `/products`

`product.create`  
Body: `name*`, `code?`, `po?`, `description?`

**事务内：** 创建产品 + 自动项目  
`name={产品名}1.0`，`code={code}-1.0`（若有且不冲突）

```json
// data
{
  "product": { "id": 1, "name": "进销存", "...": "..." },
  "defaultProject": { "id": 10, "name": "进销存1.0", "productId": 1, "status": "wait" }
}
```

### GET `/products/:id`

`product.list`  
含统计与 `default` 无关的详情。

### PUT `/products/:id`

`product.edit`  
Body: `name?`, `code?`, `po?`, `description?`, `status?`  
`status=closed` 时生效 2A。

### DELETE `/products/:id`

`product.delete`；有未删项目或需求 → 42205

### GET `/products/:id/projects`

`project.list`（或 product.list）；该产品下项目精简列表

---

## 5. 需求 Stories

### GET `/stories`

`story.list`  
Query: `productId?`, `type?`, `status?`, `assignedTo?`（`me` 表示当前用户）, `keyword?`, `sortBy?`（`type`|`pri`|`status`）, `sortOrder?`（`asc`|`desc`；与 sortBy 同时传才生效，否则按 id 倒序）

### POST `/stories`

`story.create`  
Body: `productId*`, `type?`(默认 planning), `title*`, `description?`, `pri?`, `estimate?`, `assignedTo?`  
产品 closed → 42202

### GET `/stories/:id`

`story.list`  
含 `attachments[]` 摘要、`sprints[]`（曾拉入的迭代）

### PUT `/stories/:id`

`story.edit`  
Body: `title?`, `description?`, `pri?`, `estimate?`, `assignedTo?`, `type?`, `status?`  
状态变更须符合状态机，否则 42201  
`type=story`（可交付）时仅允许改 `status`，其它字段 → 42208

### DELETE `/stories/:id`

`story.delete`；仍被迭代关联 → 42205；同时软删附件

### POST `/stories/:id/attachments`

`story.attach`  
`Content-Type: multipart/form-data`，字段名 `file`  
可交付需求禁止上传需求级附件 → 42208  
白名单与大小见 features；失败 42206

### GET `/stories/:id/remarks`

`story.list`  
返回已定稿备注列表（倒序）：`id, content, creator, createdAt, attachments[]`

### POST `/stories/:id/remarks`

`story.edit`  
创建未定稿草稿（仅可交付）；返回 `{ id, finalized:false }`；非可交付 → 42208

### POST `/stories/:id/remarks/:remarkId/finalize`

`story.edit`  
Body: `{ content }`；仅未定稿可调用一次；之后不可再改 → 42208

### DELETE `/stories/:id/remarks/:remarkId`

`story.edit`  
仅未定稿草稿可软删（提交失败清理用）；已定稿 → 42208

### POST `/stories/:id/remarks/:remarkId/attachments`

`story.attach`  
挂到备注；仅未定稿可传；备注附件不可删除

```json
// data
{
  "id": 1,
  "originalName": "a.md",
  "ext": "md",
  "sizeBytes": 1234,
  "createdAt": "..."
}
```

### GET `/attachments/:id/download`

需 Auth；校验对该 attachment 所属 story/bug 有 list 权限  
返回文件流；图片可同 URL 或另提供 `GET /attachments/:id/preview`

### DELETE `/attachments/:id`

`story.attach` 或 `bug.attach`（按 object_type）  
软删

---

## 6. 项目 Projects

### GET `/projects`

`project.list`  
Query: `productId?`, `status?`, `keyword?`

### POST `/projects`

`project.create`  
Body: `productId*`, `name*`, `code?`, `begin?`, `end?`, `pm?`, `description?`  
产品 closed → 42202

### GET `/projects/:id`

`project.list`

### PUT `/projects/:id`

`project.edit`  
不可改 `productId`；可改 status（状态机）

### DELETE `/projects/:id`

`project.delete`；有未删迭代 → 42205

### GET `/projects/:id/sprints`

`sprint.list`

### GET `/projects/:id/stories`

`story.list`  
本项目下经由任意迭代拉入过的需求（去重）

---

## 7. 迭代 Sprints

### GET `/sprints`

`sprint.list`  
Query: `projectId?`, `productId?`（经项目过滤）, `status?`

### POST `/sprints`

`sprint.create`  
Body: `projectId*`, `name*`, `begin?`, `end?`, `goal?`

### GET `/sprints/:id`

`sprint.list`  
含项目/产品摘要、`storyCount`

### PUT `/sprints/:id`

`sprint.edit`  
含 status 流转

### DELETE `/sprints/:id`

`sprint.delete`；仍有关联需求 → 42205

### GET `/sprints/:id/stories`

`sprint.list`  
已关联需求列表

### POST `/sprints/:id/stories`

`sprint.linkStory`  
Body: `{ "storyIds": [1,2,3] }`

校验：
1. 迭代 `status=doing`，否则 42203  
2. 每条需求：同产品、`type=story`、`status=active`、未在本迭代，否则 42204  
3. 写入 `sprint_story`（冗余 projectId/productId）

### DELETE `/sprints/:id/stories/:storyId`

`sprint.linkStory`  
移除关联；非 doing → 42203

### GET `/sprints/:id/story-candidates`

`sprint.linkStory`  
弹窗候选：同产品、`type=story`、`status=active`、未在本迭代  
Query: `keyword?`, page

---

## 8. 缺陷 Bugs

### GET `/bugs`

`bug.list`  
Query: `productId?`, `projectId?`, `sprintId?`, `storyId?`, `status?`, `severity?`, `pri?`, `assignedTo?`, `keyword?`, `sortBy?`（`severity`|`pri`|`status`）, `sortOrder?`（`asc`|`desc`；与 sortBy 同时传才生效，否则按 id 倒序）  
（页面从产品进时带 productId；不强制，但推荐）

### POST `/bugs`

`bug.create`  
Body: `productId*`, `title*`, `steps?`, `severity?`, `pri?`, `projectId?`, `sprintId?`, `storyId?`, `assignedTo?`  
级联一致性校验（sprint∈project∈product，story∈product）

### GET `/bugs/:id`

`bug.list`  
含 attachments

### PUT `/bugs/:id`

`bug.edit`  
可改标题步骤严重程度优先级关联指派；激活时清 resolution

### POST `/bugs/:id/resolve`

`bug.resolve`  
Body: `{ "resolution": "fixed" }` → status=resolved

### POST `/bugs/:id/close`

`bug.close` → closed（建议仅 resolved 可关，否则 42201）

### POST `/bugs/:id/activate`

`bug.edit` → active，清空 resolution/resolvedBy

### DELETE `/bugs/:id`

`bug.delete`；非 active → 42207；软删附件

### POST `/bugs/:id/attachments`

`bug.attach`  
同 story 上传

---

## 9. 工作台 Dashboard

### GET `/dashboard/summary`

需 Auth（有 dashboard 菜单）

```json
{
  "myStories": [ { "id", "title", "status", "productId" } ],
  "myBugs": [ { "id", "title", "status", "productId" } ],
  "doingSprints": [ { "id", "name", "projectId", "status" } ]
}
```

各数组最多 5 条；按用户权限过滤可见数据。

---

## 10. 资源字段字典（写入/关键读）

### Product

`id, name, code, status(normal|closed), po, description, createdBy, createdAt, updatedAt`

### Story

`id, productId, type(planning|story), title, description, pri, status(draft|active|closed), estimate, assignedTo, openedBy, createdAt, updatedAt`

### Project

`id, productId, name, code, status(wait|doing|suspended|closed), begin, end, pm, description, createdBy, createdAt, updatedAt`

### Sprint

`id, projectId, name, status(wait|doing|done|closed), begin, end, goal, createdAt, updatedAt`

### Bug

`id, productId, projectId, sprintId, storyId, title, steps, severity, pri, status(active|resolved|closed), resolution, assignedTo, openedBy, resolvedBy, createdAt, updatedAt`

### Attachment

`id, objectType(story|bug), objectId, originalName, ext, mimeType, sizeBytes, uploadedBy, createdAt`  
（不暴露 `storagePath`/`storedName` 给前端，或仅管理端可见）

---

## 11. 接口 ↔ 页面映射（摘要）

| 页面 | 主要接口 |
|------|----------|
| 登录 | `POST /auth/login` |
| 工作台 | `GET /dashboard/summary`, `GET /auth/me` |
| 产品列表/详情 | `GET/POST/PUT/DELETE /products` |
| 需求列表/详情 | `GET/POST/PUT/DELETE /stories`, attachments, remarks |
| 项目 | `CRUD /projects` |
| 迭代详情+关联 | `CRUD /sprints`, `POST/DELETE .../stories`, `.../story-candidates` |
| 缺陷 | `CRUD /bugs` + resolve/close/activate + attachments |
| 用户角色菜单 | `/users`, `/roles`, `/menus` |

---

## 12. 实现备忘

1. 建产品与默认项目须**同事务**。  
2. 附件存储目录由服务端生成 UUID 文件名；校验扩展名大小写不敏感。  
3. `assignedTo=me` 由服务端解析为当前用户 id。  
4. 所有写操作记操作者；本 V1 不做 action 日志表。  
5. CORS / CSRF 按最终前后端部署方式处理，不在本契约展开。

---

## 13. 文档状态

本契约与功能、页面对齐，可作为前后端并行开发依据。  

当前 `mvp/` 交付物：

| 文件 | 内容 |
|------|------|
| `schema.sql` | 库表 + 种子 |
| `features.md` | 功能与规则 |
| `pages.md` | 页面信息架构 |
| `api.md` | 本接口契约 |

下一步：选定技术栈并搭建工程（或先做一版 OpenAPI/Swagger YAML 导出）。
