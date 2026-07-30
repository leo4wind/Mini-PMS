# MiniPMS 功能规格（V1）

> 依据：`mvp/schema.sql`  
> 状态：已确认  
> 模型：`产品 1—N 项目 1—N 迭代`；需求/缺陷归属产品；迭代拉入需求；**不做 Task**  
> 附件：`attachment` 多态关联 story/bug  
> 建产品：自动创建同名+`1.0` 项目  
> 决策锁定：1A 管理员指定密码 · 2A 关闭产品禁建需求 · 3A 仅 doing 迭代可拉需求/建关联缺陷策略见下 · 4B 仅 active 可删缺陷 · 5A 不做自定义角色  

---

## 0. 领域模型（定稿）

```text
产品 Product
├── 需求 Story（归属产品）
│     type=planning  → 原始/规划需求
│     type=story     → 可交付需求
│     └── attachment (object_type=story)
├── 缺陷 Bug（归属产品；可关联项目/迭代/需求）
│     └── attachment (object_type=bug)
└── 项目 Project（新建产品时自动生成「{产品名}1.0」）
      └── 迭代 Sprint
            └── sprint_story → 拉入的需求
```

**不做：** Task、项目集、分支/模块表、计划/路线图表、测试单、发布、操作日志。

---

## 1. 通用规则

1. 登录鉴权；`disabled`/`deleted` 用户不可登录。  
2. 前后端均按 `menu.code` 鉴权；多角色权限取并集。  
3. 业务表软删 `deleted=1`；列表默认过滤。  
4. 人员一律 `user_id`。  
5. 内置角色不可删、code 不可改；不做新建自定义角色（5A）。  
6. 经理角色不可被清空到无法管理系统（F-SYS-08）。

### 1.1 角色矩阵

| 能力 | 开发 | 产品 | 测试 | 经理 |
|------|:----:|:----:|:----:|:----:|
| 工作台 | ✓ | ✓ | ✓ | ✓ |
| 产品 CRUD | 只读 | ✓ | | ✓ |
| 需求 CRUD | 只读 | ✓ | 只读 | ✓ |
| 项目/迭代 CRUD | 只读 | ✓（含关联需求） | 只读 | ✓ |
| 缺陷 | 编辑/解决/关闭/附件 | 只读列表 | 全量含附件 | ✓ |
| 需求附件 | | ✓ | | ✓ |
| 用户/角色/菜单 | | | | ✓ |

---

## 2. 认证

| ID | 功能 | 规则 | 权限 |
|----|------|------|------|
| F-AUTH-01 | 登录 | account+password | 匿名 |
| F-AUTH-02 | 登出 | 作废会话 | 登录用户 |
| F-AUTH-03 | 当前用户 | 资料+角色+菜单树（含 button） | 登录用户 |
| F-AUTH-04 | 修改本人密码 | 校验旧密码 | 登录用户 |

---

## 3. 系统

### 用户

| ID | 功能 | 规则 | 权限码 |
|----|------|------|--------|
| F-USER-01 | 列表 | status/关键词 | `user.list` |
| F-USER-02 | 新建 | account 唯一；**密码由管理员指定（1A）** | `user.create` |
| F-USER-03 | 编辑 | realname/email/status；account 不可改 | `user.edit` |
| F-USER-04 | 停用/启用 | disabled 建议立即失效会话 | `user.disable` |
| F-USER-05 | 分配角色 | 覆盖保存多角色 | `user.assignRole` |

### 角色 / 菜单

| ID | 功能 | 规则 | 权限码 |
|----|------|------|--------|
| F-ROLE-01 | 角色列表 | 含 builtin | `role.list` |
| F-ROLE-02 | 编辑角色 | 不可改 code；不可删 builtin | `role.edit` |
| F-ROLE-03 | 分配菜单 | role_menu | `role.assignMenu` |
| F-MENU-01~04 | 菜单树/增删改 | code 唯一；type=dir\|menu\|button | `menu.*` |
| F-SYS-08 | 经理保底 | 禁止掏空 manager 系统权限 | 系统约束 |

---

## 4. 产品

| ID | 功能 | 规则 | 权限码 |
|----|------|------|--------|
| F-PROD-01 | 列表 | | `product.list` |
| F-PROD-02 | 新建 | name 必填；code 可选唯一；**同一事务内自动创建默认项目**（见 F-PROD-06） | `product.create` |
| F-PROD-03 | 编辑 | | `product.edit` |
| F-PROD-04 | 关闭 | **closed 后禁止新建需求（2A）**；禁止新建项目 | `product.edit` |
| F-PROD-05 | 删除 | 软删；存在未删项目或需求则禁止 | `product.delete` |
| F-PROD-06 | 自动建 1.0 项目 | 新建产品成功后自动插入项目：`name = {产品名}1.0`；`product_id`=新产品；`status=wait`；`created_by`=当前用户；若产品有 code，则 `project.code = {code}-1.0`（冲突则追加短后缀或置空，实现时保证不炸） | （随 create） |

状态：`normal` ↔ `closed`

示例：产品「进销存」→ 自动项目「进销存1.0」。

---

## 5. 需求 Story

| ID | 功能 | 规则 | 权限码 |
|----|------|------|--------|
| F-STORY-01 | 列表 | 按产品；可按 type/status/指派筛 | `story.list` |
| F-STORY-02 | 新建 | 产品须 normal；默认 type=`planning` 或创建时可选；status=`draft` | `story.create` |
| F-STORY-03 | 编辑 | title/description/pri/estimate/assigned_to/type；**仅 planning**；可交付正文锁定 | `story.edit` |
| F-STORY-04 | 变更状态 | draft→active→closed；closed→active 重开；可交付后仍可改状态 | `story.edit` |
| F-STORY-05 | 转为可交付 | `planning → story`；之后正文/需求级附件不可改 | `story.edit` |
| F-STORY-06 | 删除 | 软删；已被任一迭代关联则禁止（先解绑）；级联软删其附件 | `story.delete` |
| F-STORY-07 | 上传附件 | 见 §9；planning 可挂需求；可交付后仅备注附件 | `story.attach` |
| F-STORY-08 | 删除附件 | 软删；可交付需求级附件与备注附件均不可删 | `story.attach` |
| F-STORY-09 | 下载/预览附件 | 有需求查看权即可；图片可预览，其余下载 | `story.list` |
| F-STORY-10 | 追加备注 | 仅可交付；TipTap+图/附件；定稿后不可改删 | `story.edit` / `story.attach` |

状态机：`draft → active → closed`（可重开）

---

## 6. 项目 Project

| ID | 功能 | 规则 | 权限码 |
|----|------|------|--------|
| F-PRJ-01 | 列表 | 按产品筛 | `project.list` |
| F-PRJ-02 | 新建 | **必须选所属产品**；产品须 normal | `project.create` |
| F-PRJ-03 | 编辑 | 不可更改 product_id（或仅经理可改且无迭代时）——**MVP：创建后不可改所属产品** | `project.edit` |
| F-PRJ-04 | 状态 | wait→doing→suspended/closed | `project.edit` |
| F-PRJ-05 | 删除 | 软删；有未删迭代则禁止 | `project.delete` |

说明：已取消 N:N `project_product`；一个项目只属一个产品。

---

## 7. 迭代 Sprint

| ID | 功能 | 规则 | 权限码 |
|----|------|------|--------|
| F-SPR-01 | 列表 | 按项目筛 | `sprint.list` |
| F-SPR-02 | 新建 | 挂项目下；begin/end 自定（3天/2周均可） | `sprint.create` |
| F-SPR-03 | 编辑 | name/日期/goal/status | `sprint.edit` |
| F-SPR-04 | 关联需求 | 写入 `sprint_story`；见 R-01~R-04 | `sprint.linkStory` |
| F-SPR-05 | 移除需求 | 删关联行 | `sprint.linkStory` |
| F-SPR-06 | 删除 | 软删；有关联需求则禁止 | `sprint.delete` |

状态机：`wait → doing → done → closed`

**关联需求约束（替代原「仅 doing 可建任务」）：**
- **仅 `doing` 的迭代允许关联/移除需求（延续决策 3A 精神）**  
- 需求须同属该项目的产品  
- 建议仅允许 `type=story` 且 `status=active` 的需求进入迭代（`planning` 需先转为 story）  
- **已锁定：候选列表严格仅 `type=story` + `status=active`（见 pages.md）**  
- 同一需求同一迭代不重复；同一需求可进入同一项目的后续迭代（做不完再拉）——MVP 允许；若需「一需求同时只在一个 doing 迭代」可二期加

---

## 8. 缺陷 Bug

| ID | 功能 | 规则 | 权限码 |
|----|------|------|--------|
| F-BUG-01 | 列表 | 按产品；可按项目/迭代/需求/状态筛 | `bug.list` |
| F-BUG-02 | 新建 | **product_id 必填**；project/sprint/story 可选但须一致（sprint∈project，project∈product，story∈product） | `bug.create` |
| F-BUG-03 | 编辑 | title/steps/severity/pri/关联/指派 | `bug.edit` |
| F-BUG-04 | 解决 | →resolved；resolution 必填；可填解决备注；可改指派（默认创建人）；resolved_by=当前用户 | `bug.resolve` |
| F-BUG-05 | 关闭 | resolved→closed | `bug.close` |
| F-BUG-06 | 激活 | resolved/closed→active；须填激活说明（文字/截图，写入备注）；清空 resolution/解决备注 | `bug.edit` |
| F-BUG-07 | 删除 | 软删；**仅 active 可删（4B）**；级联软删其附件 | `bug.delete` |
| F-BUG-08 | 上传附件 | 见 §9；缺陷级与备注均可挂 | `bug.attach` |
| F-BUG-09 | 删除附件 | 软删；备注附件不可删 | `bug.attach` |
| F-BUG-10 | 下载/预览附件 | 有缺陷查看权即可；图片与 mp4 可预览 | `bug.list` |
| F-BUG-11 | 追加备注 | TipTap+图/视频/附件；定稿后不可改删 | `bug.edit` / `bug.attach` |

状态机：`active → resolved → closed`（可激活回来）

---

## 9. 附件 Attachment

### 9.1 数据

表 `attachment`：多态关联，`object_type ∈ {story, bug}` + `object_id`。

| 字段 | 说明 |
|------|------|
| original_name | 用户上传时文件名 |
| stored_name | 磁盘/对象存储唯一名 |
| storage_path | 存储路径 |
| ext / mime_type / size_bytes | 元数据 |
| uploaded_by | 上传人 |

**不在 story/bug 表上建外键列**；反向查询：`WHERE object_type=? AND object_id=? AND deleted=0`。

### 9.2 允许类型（白名单）

| 类别 | 扩展名 |
|------|--------|
| Word | `doc`, `docx` |
| 文本 | `txt`, `md` |
| 图片 | `jpg`, `jpeg`, `png`, `gif`, `webp` |
| 视频 | `mp4` |

其余一律拒绝。单文件大小上限建议 **100MB**（可配置，默认见 config）。

### 9.3 功能与校验

| ID | 规则 |
|----|------|
| F-ATT-01 | 上传前校验扩展名白名单与大小 |
| F-ATT-02 | 仅当目标 story/bug 存在且未删时可挂 |
| F-ATT-03 | 删除对象时其附件一并软删 |
| F-ATT-04 | 存储路径不可被客户端指定（防路径穿越）；由服务端生成 |
| F-ATT-05 | 下载需登录且具备对应 list/查看权限 |

---

## 10. 必守规则

| ID | 规则 |
|----|------|
| R-01 | 需求/缺陷**归属产品**；迭代只「拉入」需求，不拥有需求。 |
| R-02 | 项目必须属于一个产品；`sprint.project_id` 指向该项目。 |
| R-03 | `sprint_story` 写入时：`product_id/project_id` 与 story、sprint 一致。 |
| R-04 | 仅 `doing` 迭代可关联/移除需求。 |
| R-05 | 产品 `closed` 禁止新建需求与项目。 |
| R-06 | 无 Task；执行进度靠需求状态 + 指派 + 迭代范围表达。 |
| R-07 | 新建产品必须自动创建「{产品名}1.0」项目（与产品同事务）。 |
| R-08 | 附件仅挂 story/bug；类型白名单见 §9.2。 |

---

## 11. 典型流程（进销存例子）

1. 建产品「进销存」→ **自动生成项目「进销存1.0」**  
2. 在产品下建规划需求（可上传 md/word/图片）  
3. 转为可交付需求并激活；按需再建其他项目  
4. 在「进销存1.0」下建迭代，`doing` 后拉入需求  
5. 测试提 Bug（可上传截图/说明文件），关联产品/迭代/需求  
6. 关闭 Bug/需求/迭代/项目  

---

## 12. 验收清单

- [ ] F-AUTH-01~04  
- [ ] F-USER / F-ROLE / F-MENU / F-SYS-08  
- [ ] F-PROD-01~06（含自动 1.0 项目）  
- [ ] F-STORY-01~09  
- [ ] F-PRJ-01~05  
- [ ] F-SPR-01~06  
- [ ] F-BUG-01~10  
- [ ] F-ATT-01~05  
- [ ] R-01~R-08  

---

## 13. 变更摘要（本轮）

| 变更 | 说明 |
|------|------|
| F-PROD-06 / R-07 | 新建产品自动创建 `{产品名}1.0` 项目 |
| 表 `attachment` | 多态挂 story/bug |
| 权限 | `story.attach` / `bug.attach` |
| 白名单 | doc/docx/txt/md/jpg/jpeg/png/gif/webp/mp4 |

配套：`mvp/pages.md`（页面）、`mvp/api.md`（接口）。  
下一步：选定技术栈并搭建工程。
