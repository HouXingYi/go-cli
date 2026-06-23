# ziniao

`ziniao` 是一个使用 Go 编写的命令行工具，通过 vendor-proxy 调用后端业务接口并查询 API 目录。CLI 名称为 `zn-cli`（当前版本 `0.1.0`），构建产物通常命名为 `ziniao` 或 `ziniao.exe`。

完整命令说明见 [docs/cli-api-docs.md](docs/cli-api-docs.md)。

## 命令总览

```text
zn-cli [command] [flags]

Commands:
  auth      校验本地鉴权配置
  http      通过 vendor-proxy 发送业务请求
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

构建产物会输出到 `dist/`。

## 使用方式

查看帮助：

```bash
ziniao --help
```

校验本地鉴权配置（不会向后端发请求，也不会写入磁盘）：

```bash
export CLI_AUTH_KEY=your-key
ziniao auth
```

发送 HTTP 请求（`--provider` 必填）：

```bash
export CLI_AUTH_KEY=your-key
export VENDOR_PROXY_BASE=https://api.example.com/api/v1/claw/vendor-proxy

ziniao http GET /api/user/list --provider ziniao --query '{"page":1,"pageSize":20}'
ziniao http POST /api/user/create --provider ziniao --body '{"name":"test"}'
```

查看 API 目录（未设置 `VENDOR_PROXY_BASE` 时使用内置 MockBackend）：

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

| 环境变量 | 说明 |
| --- | --- |
| `CLI_AUTH_KEY` | 鉴权密钥，HTTP 请求以 `Authorization: Bearer <key>` 发送 |
| `VENDOR_PROXY_BASE` | vendor-proxy 基础 URL；设置后对接真实后端，未设置时使用 MockBackend |

本地开发时只需 `CLI_AUTH_KEY` 即可使用 Mock 模式；沙箱环境由 `a1-browser-server` 自动注入上述变量。

内置常量（`internal/config/config.go`）：

| 常量 | 当前值 | 说明 |
| --- | --- | --- |
| `DefaultTimeout` | `10s` | HTTP 请求超时 |

## 输出格式

当前 CLI 固定使用 text 输出：成功时向 stdout 输出人类可读消息；失败时向 stderr 输出错误与提示。

## 密钥安全

`CLI_AUTH_KEY` 仅通过环境变量提供，不提供 `--token` 或命令行密钥参数。错误信息和普通输出不会回显密钥。

## 开发

运行测试：

```bash
go test ./...
```

构建：

```bash
bash scripts/build-all.sh
```

## 常见问题

| 错误信息 | 原因 | 处理建议 |
| --- | --- | --- |
| `auth key is required` | 未设置 `CLI_AUTH_KEY` | 执行 `export CLI_AUTH_KEY=...` |
| `provider is required` | `http` 未传 `--provider` | 添加 `--provider ziniao` 或 `--provider erp` |
| `body is invalid json` | `--body` 不是合法 JSON | 传入合法 JSON 对象或数组 |
| `query is invalid json` | `--query` 不是合法 JSON 对象 | 传入合法 JSON 对象，例如 `'{"page":1}'` |
| `request timed out` | 请求超时 | 检查网络连接 |
| `module/business/api "..." not found` | 目录中无对应项 | 运行 `ziniao api` 查看可用目录 |

更多细节见 [docs/cli-api-docs.md](docs/cli-api-docs.md)。
