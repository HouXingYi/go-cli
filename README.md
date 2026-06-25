# ziniao

`ziniao` 是一个使用 Go 编写的命令行工具族，通过 vendor-proxy 调用后端业务接口并查询 API 目录。当前版本 `0.1.0`。

完整命令说明见 [docs/cli-api-docs.md](docs/cli-api-docs.md)。

## CLI 产物

| 产物 | 入口 | 状态目录 | 内置 cli-type |
| --- | --- | --- | --- |
| `zn-ent` | `./cmd/zn-ent` | `UserConfigDir()/zn-ent/` | `ent`（占位，后端定稿后可改） |
| `zn-eco` | `./cmd/zn-eco` | `UserConfigDir()/zn-eco/` | `eco`（占位，后端定稿后可改） |

`zn-eco` 子命令与 flags 与 `zn-ent` 完全相同；下文示例默认使用 `zn-ent`，使用时将命令名替换为 `zn-eco` 即可。

## 命令总览

```text
zn-ent [command] [flags]

Commands:
  agent     Agent-oriented CLI guidance
  auth      校验本地鉴权配置
  config    管理 CLI 配置（默认 module 等）
  http      通过 vendor-proxy 发送业务请求
  api       查看 HTTP API 目录
  version   查看 CLI 版本
```

不带子命令运行时显示帮助信息。

## Agent 使用

Agent 可通过内置指南自举加载使用规范（无需鉴权、离线可用）：

```bash
zn-ent agent guide
```

输出为按当前变体渲染的 Markdown（模板 `skills/shared/SKILL.md`），可直接放入 Agent 上下文。推荐工作流见指南中的 **推荐工作流** 章节；业务 API 细节请用 `zn-ent api` 动态发现，勿依赖静态 API 列表。

## 从源码构建

构建前需要先安装 Go，并确保终端中可以执行 `go version`。

构建当前平台：

```bash
go build -o bin/zn-ent ./cmd/zn-ent
go build -o bin/zn-eco ./cmd/zn-eco
```

Windows：

```bash
go build -o bin/zn-ent.exe ./cmd/zn-ent
go build -o bin/zn-eco.exe ./cmd/zn-eco
```

一次构建多个平台（`dist/zn-ent-*`、`dist/zn-eco-*`）：

Windows：

```bat
scripts\build-all.bat
```

也可以使用 PowerShell：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-all.ps1
```

Linux / macOS / Git Bash：

```bash
bash scripts/build-all.sh
```

构建产物会输出到 `dist/`。变体清单见 [build/variants.yaml](build/variants.yaml)。

## 使用方式

查看帮助：

```bash
zn-ent --help
```

校验本地鉴权配置（不会向后端发请求，也不会写入磁盘）：

```bash
export CLI_AUTH_KEY=your-key
zn-ent auth
```

推荐工作流：先查看 API 目录，设置默认大模块（module），再发送请求：

```bash
export CLI_AUTH_KEY=your-key
export VENDOR_PROXY_BASE=https://api.example.com/api/v1/claw/vendor-proxy

zn-ent api
zn-ent config module set ziniao
zn-ent http GET /api/user/list --query '{"page":1,"pageSize":20}'
zn-ent http POST /api/user/create --body '{"name":"test"}'

# 临时切换大模块
zn-ent http GET /api/order/list --module erp
```

查看 API 目录（未设置 `VENDOR_PROXY_BASE` 时使用内置 MockBackend）：

```bash
zn-ent api
zn-ent api ziniao user
zn-ent api ziniao user list
```

查看版本：

```bash
zn-ent version
```

## 配置

| 环境变量 | 说明 |
| --- | --- |
| `CLI_AUTH_KEY` | 鉴权密钥，HTTP 请求以 `Authorization: Bearer <key>` 发送 |
| `VENDOR_PROXY_BASE` | vendor-proxy 基础 URL；设置后对接真实后端，未设置时使用 MockBackend |
| `ZINIAO_MODULE` | 默认一级大模块；优先级低于 `http --module`，高于 `config module set` 持久化值 |
| `ZINIAO_CONFIG_DIR` | CLI 状态目录；默认 `UserConfigDir()/<AppName>/`（`zn-ent` 或 `zn-eco`） |

本地开发时只需 `CLI_AUTH_KEY` 即可使用 Mock 模式；沙箱环境由 `a1-browser-server` 自动注入上述变量。

远程请求还会自动携带编译期内置的 `Cli-Type` 请求头（值因 `zn-ent` / `zn-eco` 而异）。

内置常量（`internal/config/config.go`）：

| 常量 | 当前值 | 说明 |
| --- | --- | --- |
| `DefaultTimeout` | `10s` | HTTP 请求超时 |

## 输出格式

当前 CLI 固定使用 text 输出：成功时向 stdout 输出人类可读消息；失败时向 stderr 输出错误与提示。

## 密钥安全

`CLI_AUTH_KEY` 仅通过环境变量提供，不提供 `--token` 或命令行密钥参数。错误信息和普通输出不会回显密钥。

## 开发

### 前置条件

- Go **1.22+**（`go.mod`）
- 在仓库根目录执行；Windows 可用 Git Bash / PowerShell

### 自动化测试（提交前必跑）

| 场景 | 命令 |
| --- | --- |
| 全量 | `go test ./...` |
| 看用例名 | `go test -v ./...` |
| 只测 CLI | `go test ./internal/cli/...` |
| 单测 | `go test ./internal/cli -run TestHTTPCommandUsesMockBackend` |
| 禁用缓存 | `go test -count=1 ./...` |

各包覆盖重点：

| 包 | 覆盖重点 |
| --- | --- |
| `internal/cli` | 各子命令端到端：auth / api / http / config / version / agent guide |
| `internal/backend` | MockBackend 目录与代理；HTTPBackend 用 `httptest` 验请求格式与 envelope 解析 |
| `internal/config` | 状态文件读写、`module` 优先级（flag > env > 持久化） |
| `internal/catalog` | 目录 JSON 解析 |
| `internal/agent` | 嵌入 SKILL 模板非空、按 variant 渲染 AppName |
| `internal/output` | text/json 成功与错误输出格式 |

### 本地手动冒烟（Mock 模式，无需后端）

未设置 `VENDOR_PROXY_BASE` 时 CLI 走内置 MockBackend（见上文「配置」）；`CLI_AUTH_KEY` 任意非空值即可。

快速迭代：

```bash
go run ./cmd/zn-ent version
go run ./cmd/zn-eco version
```

隔离配置（避免污染本机状态目录）：

Git Bash / Linux / macOS：

```bash
export CLI_AUTH_KEY=dev
export ZINIAO_CONFIG_DIR="$(mktemp -d)"
unset VENDOR_PROXY_BASE
```

PowerShell：

```powershell
$env:CLI_AUTH_KEY = "dev"
$env:ZINIAO_CONFIG_DIR = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path "$($_.FullName)-zn-ent" }
Remove-Item Env:VENDOR_PROXY_BASE -ErrorAction SilentlyContinue
```

冒烟检查清单（按顺序跑，预期均 exit 0）：

```bash
go run ./cmd/zn-ent version
go run ./cmd/zn-ent auth
go run ./cmd/zn-ent api
go run ./cmd/zn-ent api ziniao user list
go run ./cmd/zn-ent config module set ziniao
go run ./cmd/zn-ent config module get
go run ./cmd/zn-ent http GET /api/user/list --query '{"page":1}'
go run ./cmd/zn-ent agent guide | head -5
```

负面用例（错误提示预期见下文「常见问题」表）：

```bash
unset CLI_AUTH_KEY
go run ./cmd/zn-ent auth

export CLI_AUTH_KEY=dev
go run ./cmd/zn-ent http GET /api/user/list --query 'not-json'
```

构建二进制后，将 `go run ./cmd/zn-ent` 换成 `bin/zn-ent` 或 `bin/zn-ent.exe` 做同样检查。

### 联调真实后端（可选）

仅在需要验证 HTTPBackend 或沙箱对接时执行：

```bash
export CLI_AUTH_KEY=your-key
export VENDOR_PROXY_BASE=https://api.example.com/api/v1/claw/vendor-proxy
go run ./cmd/zn-ent api
go run ./cmd/zn-ent http GET /api/user/list --module ziniao --query '{"page":1,"pageSize":20}'
```

协议细节见 [docs/cli-integration.md](docs/cli-integration.md)；命令参数见 [docs/cli-api-docs.md](docs/cli-api-docs.md)。

### 构建验证

详见上文「从源码构建」。发布前建议：

```bash
go build -o bin/zn-ent ./cmd/zn-ent
go build -o bin/zn-eco ./cmd/zn-eco
bash scripts/build-all.sh
```

构建后用 `bin/zn-ent` 再跑一遍冒烟检查清单。

### 按改动类型的自测清单

| 改动范围 | 建议动作 |
| --- | --- |
| CLI 命令 / 参数 / 错误文案 | `go test ./internal/cli/...` + Mock 冒烟清单 |
| Mock 目录或模拟响应 | `go test ./internal/backend/...` + `zn-ent api` / `zn-ent http` 手动确认 |
| HTTP 请求路径、鉴权头、envelope | `go test ./internal/backend/...`；有环境则加真实后端联调 |
| `config` / 状态文件 | `go test ./internal/config/...`；冒烟 `config module set/get/clear` |
| `skills/shared/SKILL.md` | `go test ./internal/agent/...`（嵌入模板与按 variant 渲染） |
| 输出格式 | `go test ./internal/output/...` |

### 推荐工作流

改代码 → `go test ./...` → Mock 冒烟 →（如涉及后端）联调 → `go build` 验证二进制。

## 常见问题

| 错误信息 | 原因 | 处理建议 |
| --- | --- | --- |
| `auth key is required` | 未设置 `CLI_AUTH_KEY` | 执行 `export CLI_AUTH_KEY=...` |
| `module is required` | `http` 未配置默认 module 且未传 `--module` | 执行 `zn-ent config module set ziniao` 或添加 `--module ziniao` |
| `body is invalid json` | `--body` 不是合法 JSON | 传入合法 JSON 对象或数组 |
| `query is invalid json` | `--query` 不是合法 JSON 对象 | 传入合法 JSON 对象，例如 `'{"page":1}'` |
| `request timed out` | 请求超时 | 检查网络连接 |
| `module/business/api "..." not found` | 目录中无对应项 | 运行 `zn-ent api` 查看可用目录 |

更多细节见 [docs/cli-api-docs.md](docs/cli-api-docs.md)。
