# ziniao

`ziniao` 是一个使用 Go 编写的、基于 token 调用 HTTP 接口的命令行工具。CLI 名称为 `zn-cli`（当前版本 `0.1.0`），构建产物通常命名为 `ziniao` 或 `ziniao.exe`。项目将命令解析、运行时配置、应用用例、HTTP 通信和输出格式化分层处理，方便后续新增命令和接口能力。

完整命令说明见 [docs/cli-api-docs.md](docs/cli-api-docs.md)。

## 命令总览

```text
zn-cli [command] [flags]

Commands:
  auth      校验本地鉴权配置
  http      发送带 Bearer Token 的 HTTP 请求
  api       查看 HTTP API 目录
  version   查看 CLI 版本
```

不带子命令运行时显示帮助信息。

## 从源码构建

构建前需要先安装 Go，并确保终端中可以执行 `go version`。

构建当前平台：

```bash
go build -o bin/ziniao ./cmd/ziniao
```

Windows：

```bash
go build -o bin/ziniao.exe ./cmd/ziniao
```

一次构建多个平台：

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

构建产物会输出到 `dist/`：

```text
dist/ziniao-linux-amd64
dist/ziniao-linux-arm64
dist/ziniao-darwin-amd64
dist/ziniao-darwin-arm64
dist/ziniao-windows-amd64.exe
```

Linux 环境中使用前需要赋予执行权限：

```bash
chmod +x dist/ziniao-linux-amd64
```

## 使用方式

查看帮助：

```bash
ziniao --help
```

校验本地鉴权配置（不会向后端发请求，也不会写入磁盘）：

```bash
export ZINIAO_TOKEN=your-token
ziniao auth
```

发送 HTTP 请求（请求发往内置 API 基础地址 `https://gateway.ziniao.com`）：

```bash
export ZINIAO_TOKEN=your-token
ziniao http GET /api/user/list --query page=1 --query pageSize=20
ziniao http POST /api/user/create --body '{"name":"test"}'
```

查看 API 目录（当前使用内置 Mock 数据，不依赖 token）：

```bash
ziniao api
ziniao api ziniao user
ziniao api ziniao user list
```

查看版本：

```bash
ziniao version
```

## 配置

当前唯一用户可配置项为环境变量 `ZINIAO_TOKEN`，通过 Viper 读取（`internal/config` 保留 Viper 基础设施，便于后续扩展更多配置项）。

| 环境变量 | 说明 |
| --- | --- |
| `ZINIAO_TOKEN` | 访问 token，HTTP 请求以 `Authorization: Bearer <token>` 发送 |

以下值当前为内置常量（定义于 `internal/config/config.go`），不可通过命令行或环境变量修改：

| 常量 | 当前值 | 说明 |
| --- | --- | --- |
| `DefaultBaseURL` | `https://gateway.ziniao.com` | HTTP API 基础地址 |
| `DefaultTimeout` | `10s` | HTTP 请求超时 |

## 输出格式

当前 CLI 固定使用 text 输出：成功时向 stdout 输出人类可读消息；失败时向 stderr 输出错误与提示。

## Token 安全

token 仅通过环境变量 `ZINIAO_TOKEN` 提供，不提供 `--token` 命令行参数。错误信息和普通输出不会回显 token。

## 开发

运行测试：

```bash
go test ./...
```

构建：

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

## 常见问题

| 错误信息 | 原因 | 处理建议 |
| --- | --- | --- |
| `token is required` | 未设置 `ZINIAO_TOKEN` | 执行 `export ZINIAO_TOKEN=...` |
| `path is invalid` | `http` 命令 path 不以 `/` 开头 | 使用 `/api/...` 形式 |
| `body is invalid json` | `--body` 不是合法 JSON | 传入合法 JSON 对象或数组 |
| `invalid key=value pair` | `--query` 或 `--header` 格式错误 | 使用 `key=value` 格式 |
| `request timed out` | 请求超时 | 检查网络连接 |
| `module/business/api "..." not found` | 目录中无对应项 | 运行 `ziniao api` 查看可用目录 |

HTTP 401/403 时错误 kind 为 `auth`，其他非 2xx 为 `api`。更多细节见 [docs/cli-api-docs.md](docs/cli-api-docs.md)。
