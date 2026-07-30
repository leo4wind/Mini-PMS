# MiniPMS 部署产物

本目录存放后端交叉编译与打包脚本。全栈 Docker Compose 在平级目录 **[`../compose`](../compose)**。

## 推荐：Docker 全栈

```powershell
cd mvp\deploy
.\pack-release.ps1
cd ..\compose
docker compose down -v
docker compose up -d --build
```

浏览器：**http://127.0.0.1/mini-pms/**  
账号：`admin` / `123456`  
说明见 [`../compose/README.md`](../compose/README.md)。

## 目录结构

```
mvp/
├── deploy/                   # 本目录：交叉编译 / 打 release
│   ├── README.md
│   ├── build-linux.ps1
│   ├── pack-release.ps1      # → 输出到 ../compose/release/
│   └── linux-amd64/          # 裸机用后端包
└── compose/                  # 与 deploy 平级：Docker Compose
    ├── docker-compose.yml
    ├── Dockerfile.api
    ├── nginx.conf
    ├── config.yaml
    ├── README.md
    └── release/
```

## 仅打包 Linux 后端（无 Docker）

```powershell
cd mvp/deploy
.\build-linux.ps1
# ARM: .\build-linux.ps1 -Arch arm64
```

或手动：

```powershell
cd mvp/backend
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
mise exec go -- go build "-ldflags=-s -w" -o ../deploy/linux-amd64/minipms-server ./cmd/server
Copy-Item configs/config.yaml ../deploy/linux-amd64/configs/config.yaml -Force
```

### 在 Linux 上直接运行后端

1. 拷贝整个 `linux-amd64/`，改 `configs/config.yaml`。
2. 导入 `mvp/schema.sql`。
3. `chmod +x minipms-server && ./minipms-server`

可选：`export MINIPMS_CONFIG=/etc/minipms/config.yaml`  
默认 `:8088`。健康检查：`GET http://127.0.0.1:8088/health`。

## 注意

- Windows 不要直接运行 `minipms-server`（Linux ELF）。
- Compose 数据库初始化依赖空数据卷 + `schema.sql`；见 compose README。
