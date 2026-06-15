# ziniao CLI 使用文档

`ziniao` 是一个基于 Bearer Token 调用 HTTP API 的命令行工具。当前 CLI 提供认证校验和执行请求流程两个命令，并支持通过命令行参数、环境变量和配置文件传入运行配置。

## 快速开始

先确认可执行文件可以正常运行：

```bash
ziniao --help
```

验证 token 是否有效：

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

如需 JSON 输出：

```bash
ziniao request run \
  --base-url https://api.example.com \
  --token your-token \
  --output json
```

## 命令总览

```text
ziniao [command] [flags]

Commands:
  auth verify   Verify whether the token is valid
  request run   Run the configured HTTP request
```

不带子命令运行 `ziniao` 时会显示帮助信息。

## 全局参数

所有子命令都支持以下全局参数：

```text
--config <path>       配置文件路径，默认查找当前目录或用户主目录下的 .ziniao.yaml
--token <token>       访问 token，会以 Authorization: Bearer <token> 发送
--base-url <url>      HTTP API 基础地址，例如 https://api.example.com
--timeout <duration>  HTTP 请求超时时间，默认 10s，例如 500ms、10s、1m
--output <format>     输出格式，支持 text 或 json，默认 text
--verbose             启用更详细日志
```

配置优先级从高到低为：

1. 命令行参数。
2. 环境变量。
3. 配置文件。
4. 默认值。

## 配置文件

默认情况下，CLI 会查找当前目录和用户主目录中的 `.ziniao.yaml`。也可以通过 `--config` 指定其它配置文件：

```bash
ziniao auth verify --config ./ziniao.yaml
```

配置文件示例：

```yaml
base_url: "https://api.example.com"
token: "your-token"
timeout: "10s"
output: "text"
verbose: false
```

## 环境变量

也可以使用 `ZINIAO_` 前缀的环境变量传入配置：

```bash
export ZINIAO_BASE_URL=https://api.example.com
export ZINIAO_TOKEN=your-token
export ZINIAO_TIMEOUT=10s
export ZINIAO_OUTPUT=json
export ZINIAO_VERBOSE=false
```

Windows PowerShell 示例：

```powershell
$env:ZINIAO_BASE_URL = "https://api.example.com"
$env:ZINIAO_TOKEN = "your-token"
$env:ZINIAO_TIMEOUT = "10s"
$env:ZINIAO_OUTPUT = "json"
```

## `auth verify`

`auth verify` 用于校验当前 token 是否可用。

```bash
ziniao auth verify \
  --base-url https://api.example.com \
  --token your-token
```

该命令会发送请求：

```http
GET /auth/verify
Authorization: Bearer <token>
Accept: application/json
```

如果接口响应中没有 `message` 字段，CLI 会使用默认成功文案：

```text
Auth verified successfully.
```

## `request run`

`request run` 用于执行当前配置的 HTTP 请求流程。

```bash
ziniao request run \
  --base-url https://api.example.com \
  --token your-token
```

该命令会发送请求：

```http
GET /request/run
Authorization: Bearer <token>
Accept: application/json
```

如果接口响应中没有 `message` 字段，CLI 会使用默认成功文案：

```text
Request completed successfully.
```

## 输出格式

默认 `text` 格式只输出成功消息或错误信息：

```text
Request completed successfully.
```

错误时会输出错误和提示：

```text
Error: token is required
Hint: pass --token or set ZINIAO_TOKEN.
```

`json` 格式会输出结构化结果，适合脚本或自动化场景：

```json
{
  "success": true,
  "data": {
    "message": "Request completed successfully."
  }
}
```

错误时：

```json
{
  "success": false,
  "error": {
    "message": "token is required",
    "hint": "pass --token or set ZINIAO_TOKEN.",
    "kind": "config"
  }
}
```

## 常见问题

### `token is required`

没有提供 token。可以通过 `--token`、`ZINIAO_TOKEN` 或配置文件中的 `token` 提供。

### `base_url is required`

没有提供 API 基础地址。可以通过 `--base-url`、`ZINIAO_BASE_URL` 或配置文件中的 `base_url` 提供。

### `base_url is invalid`

`base_url` 必须是完整 URL，例如：

```text
https://api.example.com
```

### `output format is invalid`

`--output` 只支持 `text` 或 `json`。

### `unauthorized` 或权限相关错误

token 可能无效、已过期，或缺少访问目标接口的权限。请确认 token 状态和接口权限。

### `request timed out`

请求超时。可以增加超时时间，或检查网络和 API 服务状态：

```bash
ziniao request run \
  --base-url https://api.example.com \
  --token your-token \
  --timeout 30s
```

## 当前 API 说明

在真实 API 契约确认前，当前 CLI 使用以下占位接口：

```text
GET /auth/verify
GET /request/run
```

两个接口都使用 Bearer Token 鉴权：

```http
Authorization: Bearer <token>
```
