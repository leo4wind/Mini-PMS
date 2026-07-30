# MiniPMS Docker Compose（WSL / Docker Desktop）

与 [`../deploy`](../deploy) 平级。用 Docker Desktop（WSL2）一键拉起：**Nginx + API + MySQL**。

```
Browser
  http://127.0.0.1/mini-pms/     → Nginx 静态前端
  http://127.0.0.1/mini-pms/api/ → API :8088
                                 → MySQL :3306（DDL+种子：initdb/01-schema.sql）
```

## 前置条件

1. Docker Desktop 已启动（WSL2 engine）。
2. 本机 80 端口可用。
3. 已生成产物到 `compose/release/`（见下方）。

## 1. 打包产物

在 Windows PowerShell：

```powershell
cd mvp\deploy
.\pack-release.ps1
```

生成：

```
mvp/compose/release/
├── minipms-server
└── web/                 # Vite dist（base=/mini-pms/）
```

同时会把 `mvp/schema.sql` 同步到 `compose/initdb/01-schema.sql`。

## 2. 启动（务必注意数据库初始化）

请在仓库目录操作（不要用其它拷贝路径）：

```bash
cd /mnt/e/lcd/code/zentaopms/mvp/compose
ls -la initdb/01-schema.sql    # 必须是普通文件（-rw...），不能是目录（drw...）
docker compose down -v
docker compose up -d --build
```

`initdb/01-schema.sql`（DDL + admin 种子）只在 **MySQL 数据卷为空时** 导入一次。`-v` 会清空卷以便重新导入。

验证：

```bash
docker compose exec mysql ls -la /docker-entrypoint-initdb.d/
# 应显示普通文件，不是目录
docker compose exec mysql mysql -uroot -pminipms -e "USE minipms; SHOW TABLES; SELECT account FROM user;"
```

## 3. 访问

| 项 | 地址 |
|----|------|
| 前端 | **http://127.0.0.1/mini-pms/** |
| 健康检查 | http://127.0.0.1/mini-pms/health |
| 账号 | `admin` / `123456` |

根路径 `/` 留给其他站点或文件；本应用只占用 `/mini-pms/`。

## 常用命令

```bash
docker compose ps
docker compose logs -f mysql
docker compose logs -f api
docker compose exec mysql mysql -uroot -pminipms -e "USE minipms; SHOW TABLES; SELECT account FROM user;"
docker compose down          # 停服务，保留数据
docker compose down -v       # 停服务并清空库（下次 up 会重新跑 schema）
```

## 排错

| 现象 | 处理 |
|------|------|
| `user` 表不存在 / `input source is a directory` | schema 被挂成了目录。确认 `initdb/01-schema.sql` 是文件后 `down -v && up -d --build` |
| 路径不对 | 必须用仓库 `mvp/compose`；`~/projects/deploy/compose` 这类拷贝容易缺 `initdb` |
| `mise ... not trusted` | `mise trust /mnt/e/lcd/code/zentaopms/mvp`（与 Docker 无关） |
| 拉镜像失败 | Docker Desktop → Docker Engine 加 `registry-mirrors` |
| 404 / JS 路径错 | 重新 `pack-release.ps1`，确认 `release/web/index.html` 存在 |
| 空白页但 HTML 有 | 强刷缓存；访问 `/mini-pms/` 不是 `/` |
