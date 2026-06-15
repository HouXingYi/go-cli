# ziniao

`ziniao` 是一个使用 Go 编写的、基于 token 调用 HTTP 接口的命令行工具。项目将命令解析、运行时配置、应用用例、HTTP 通信和输出格式化分层处理，方便后续新增命令和接口能力。

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

验证 token：

```bash
ziniao auth verify \
  --base-url https://api.example.com \
  --token your-token
```

执行当前请求流程：

```bash
ziniao request run \
  --base-url https://api.example.com \
  --token your-token
```

输出 JSON：

```bash
ziniao auth verify \
  --base-url https://api.example.com \
  --token your-token \
  --output json
```

## 配置

配置优先级如下：

1. 命令行参数。
2. 环境变量。
3. 配置文件。
4. 默认值。

全局参数：

```bash
--config <path>       配置文件路径
--token <token>       访问 token
--base-url <url>      HTTP API 基础地址
--timeout <duration>  HTTP 请求超时时间，默认 10s
--output <format>     输出格式，支持 text 或 json
--verbose             启用更详细日志
```

环境变量：

```bash
ZINIAO_BASE_URL=https://api.example.com
ZINIAO_TOKEN=your-token
ZINIAO_TIMEOUT=10s
ZINIAO_OUTPUT=json
```

配置文件示例：

```yaml
base_url: "https://api.example.com"
token: ""
timeout: "10s"
output: "text"
```

默认情况下，CLI 会在当前目录和用户主目录查找 `.ziniao.yaml`。也可以通过 `--config` 显式指定配置文件路径。

## Token 安全

当前版本不会主动把 token 写入本地配置。一次性使用时可以通过 `--token` 传入，本地自动化场景可以使用 `ZINIAO_TOKEN`。错误信息和普通输出不应回显完整 token。

## 当前 API 占位

在真实 API 契约确认前，CLI 暂时使用以下占位接口：

```text
GET /auth/verify
GET /request/run
```

两个请求都使用 Bearer Token 鉴权：

```http
Authorization: Bearer <token>
```

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

### token is required

通过 `--token` 传入 token，或设置 `ZINIAO_TOKEN`。

### base_url is required

通过 `--base-url` 传入接口基础地址，或设置 `ZINIAO_BASE_URL`。

### unauthorized

token 可能无效、已过期，或缺少访问目标接口的权限。

### request timed out

可以增加超时时间：

```bash
ziniao request run \
  --base-url https://api.example.com \
  --token your-token \
  --timeout 30s
```
