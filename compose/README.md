# MiniPMS Docker Compose（WSL / Docker Desktop）

与 [`../deploy`](../deploy) 平级。用 Docker Desktop（WSL2）一键拉起：**Nginx + API + MySQL**。

```
Browser
  http://127.0.0.1/mini-pms/     → Nginx 静态前端
  http://127.0.0.1/mini-pms/api/ → API :8088
                                 → MySQL（数据在 ./data/mysql）
                                 → 附件（./data/uploads）
```

持久化目录（相对本 `compose/`，可随目录一起拷到服务器 / WSL home）：

| 宿主机路径 | 内容 |
|------------|------|
| `data/mysql/` | MySQL 数据文件 |
| `data/uploads/` | 上传附件 |

WSL 示例：`/home/leo/projects/compose/data/...`  
Windows 资源管理器：`\\wsl.localhost\Ubuntu\home\leo\projects\compose\data\...`

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
# 或你的部署目录，例如：cd ~/projects/compose
mkdir -p data/mysql data/uploads
ls -la initdb/01-schema.sql    # 必须是普通文件（-rw...），不能是目录（drw...）
docker compose down
docker compose up -d --build
```

`initdb/01-schema.sql`（DDL + admin 种子）只在 **`data/mysql` 为空时** 导入一次。若要强制重装库：先停栈，删掉 `data/mysql` 内容后再 `up`。

`docker compose down` **不会**删 `./data/`；以前用过命名卷的，可用 `docker volume rm compose_minipms-mysql-data compose_minipms-uploads` 清掉旧卷（与新的目录挂载无关）。

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
docker compose down          # 停服务；./data/ 仍保留
# 清空库/附件：rm -rf data/mysql/* data/uploads/* 后再 up（会重新跑 schema）
```

## 排错

| 现象 | 处理 |
|------|------|
| `user` 表不存在 / `input source is a directory` | schema 被挂成了目录。确认 `initdb/01-schema.sql` 是文件；若 `data/mysql` 已有空库，清空该目录后再 `up` |
| 路径不对 | 必须用仓库 `mvp/compose`；`~/projects/deploy/compose` 这类拷贝容易缺 `initdb` |
| `mise ... not trusted` | `mise trust /mnt/e/lcd/code/zentaopms/mvp`（与 Docker 无关） |
| 拉镜像失败 | Docker Desktop → Docker Engine 加 `registry-mirrors` |
| 404 / JS 路径错 | 重新 `pack-release.ps1`，确认 `release/web/index.html` 存在 |
| 空白页但 HTML 有 | 强刷缓存；访问 `/mini-pms/` 不是 `/` |
