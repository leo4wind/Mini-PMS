# MiniPMS

最小研发管理系统。规格见同目录文档；本目录为可运行工程骨架。

## 技术栈

| 端 | 技术 |
|----|------|
| 后端 | Go + Gin + GORM + MySQL + JWT |
| 前端 | Vue 3 + Vite + Naive UI + Pinia + Vue Router |

## 文档

- `schema.sql` — 数据库
- `features.md` — 功能规格
- `pages.md` — 页面信息架构
- `api.md` — 接口契约

## 当前已实现

- 后端：认证；产品（新建自动 `{产品名}1.0` 项目）；项目；**需求 / 迭代(+关联) / 缺陷 / 附件 / 工作台汇总**；用户/角色/菜单；JWT + 菜单权限
- 前端：全业务页（产品、项目、需求、迭代、缺陷、附件面板、工作台）+ 系统管理

## 建议开发顺序

1. ~~产品~~ ✅  
2. ~~用户 / 角色 / 菜单（含关联）~~ ✅  
3. ~~项目~~ ✅  
4. ~~需求~~ ✅  
5. ~~迭代 + 关联需求~~ ✅  
6. ~~缺陷~~ ✅  
7. ~~附件~~ ✅  
8. ~~工作台汇总~~ ✅  

MVP 规格内功能已全部落地。

## 环境准备

1. 安装 **mise**（已有可跳过），在本目录执行：

```bash
cd mvp
mise trust
mise install   # 安装 go 1.24.4 + node 24
```

2. MySQL 8（当前配置 `root / 1234 @ 127.0.0.1:3306`），导入库表：

```bash
# 若有 mysql 客户端：
mysql -uroot -p1234 < schema.sql
# 或用任意客户端执行 schema.sql
```

3. 默认账号：
   - 账号：`admin`
   - 密码：`password`（schema 占位哈希；生产务必修改）

生成新密码哈希：

```bash
cd backend
mise exec go -- go run ./cmd/hashpwd your-password
```

4. DSN 已写在 `backend/configs/config.yaml`。国内 Go 代理已写入 `mise.toml`（GOPROXY=goproxy.cn）。

## 启动后端

```bash
cd mvp/backend
mise exec go -- go run ./cmd/server
```

默认监听 `:8088`（避免与本机已占用的 8080 冲突）。健康检查：`GET http://127.0.0.1:8088/health`

## 启动前端

```bash
cd mvp/frontend
mise exec node -- npm install
mise exec node -- npm run dev
```

浏览器打开 Vite 提示的地址（通常 `http://127.0.0.1:5173`），API 经代理转发到 `:8088`。

## 附件

上传目录默认 `backend/uploads/`（可在 `configs/config.yaml` 的 `upload.dir` 修改）。白名单：doc/docx/txt/md + 图片；单文件上限 20MB。
