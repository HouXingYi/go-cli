# zn-cli 文档

## CLI 名称

`zn-cli`（`internal/config.AppName`，当前版本 `0.1.0`）

可执行文件入口：`cmd/ziniao/main.go`，构建产物通常命名为 `ziniao` 或 `ziniao.exe`。

## 命令总览

```text
zn-cli [command] [flags]

Commands:
  auth      Configure CLI authentication
  http      Send an authenticated HTTP request
  api       Inspect backend HTTP API catalog
  version   Print CLI version
```

不带子命令运行 `zn-cli` 时显示帮助信息。

## 全局参数

所有子命令共享以下全局参数（定义于 `internal/cli/root.go`）：

| 参数 | 环境变量 | 说明 |
| --- | --- | --- |
| `--config <path>` | — | 配置文件路径；未指定时查找当前目录和用户主目录下的 `.ziniao.yaml` |
| `--token <token>` | `ZINIAO_TOKEN` | 访问 token，HTTP 请求以 `Authorization: Bearer <token>` 发送 |
| `--base-url <url>` | `ZINIAO_BASE_URL` | HTTP API 基础地址，须为完整 URL，例如 `https://api.example.com` |
| `--timeout <duration>` | `ZINIAO_TIMEOUT` | HTTP 请求超时，默认 `10s`，支持 `500ms`、`10s`、`1m` 等 |
| `--output <format>` | `ZINIAO_OUTPUT` | 输出格式：`text`（默认）或 `json` |
| `--verbose` | `ZINIAO_VERBOSE` | 启用详细日志（当前已注册，尚未在业务逻辑中消费） |

配置优先级从高到低：命令行参数 → 环境变量 → 配置文件 → 默认值。

配置文件示例（`.ziniao.yaml`）：

```yaml
base_url: "https://api.example.com"
token: "your-token"
timeout: "10s"
output: "text"
verbose: false
```

## 输出格式

### text（默认）

成功时向 stdout 输出人类可读消息；失败时向 stderr 输出错误与提示：

```text
Error: token is required
Hint: pass --token or set ZINIAO_TOKEN.
```

### json

成功时输出结构化信封：

```json
{
  "success": true,
  "data": { }
}
```

失败时：

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

错误 `kind` 取值：`config`、`network`、`auth`、`api`、`output`。

---

## 核心命令

### `zn-cli auth`

**说明：** 校验本地鉴权配置是否完整（`token` 与 `base_url` 均已提供且 `base_url` 格式合法）。当前实现**不会**向后端发送请求，也**不会**将配置写入磁盘。

**命令：**

```bash
zn-cli auth
```

**参数：** 无（继承全局参数）。

**前置条件：**

- `token` 非空
- `base_url` 为合法的绝对 URL

**输出：**

text 模式：

```text
Authentication configured successfully.
```

json 模式 `data` 字段：

```json
{
  "baseUrl": "https://api.example.com",
  "status": "configured"
}
```

**示例：**

```bash
zn-cli auth --base-url https://api.example.com --token your-token
```

---

### `zn-cli http`

**说明：** 向 `base_url` 拼接后的目标路径发送带 Bearer Token 的 HTTP 请求，并打印响应。

**命令：**

```bash
zn-cli http <method> <path> [flags]
```

**参数：**

| 参数 | 说明 |
| --- | --- |
| `<method>` | HTTP 方法，例如 `GET`、`POST`、`PUT`、`DELETE`（大小写不敏感，发送前会转为大写） |
| `<path>` | 后端接口路径，须以 `/` 开头，例如 `/api/user/list` |

**命令级 flags：**

| 参数 | 说明 |
| --- | --- |
| `--query <key=value>` | 追加 URL 查询参数，可重复传入 |
| `--body <json>` | 请求体，须为合法 JSON，适用于 `POST`、`PUT` 等 |
| `--header <key=value>` | 追加请求头，可重复传入 |

**请求行为：**

- 目标 URL：`base_url` + `path` + query 参数
- 自动设置 `Accept: application/json`
- 有 body 时自动设置 `Content-Type: application/json`（除非已通过 `--header` 指定）
- 自动设置 `Authorization: Bearer <token>`

**前置条件：** 需要有效的 `token` 与 `base_url`。

**输出：**

text 模式打印格式化后的响应内容：

- 响应体为合法 JSON 时，缩进输出 JSON
- 响应体为空时，输出 HTTP 状态行（例如 `200 OK`）
- 否则原样输出响应体文本

json 模式 `data` 字段为响应结构：

```json
{
  "statusCode": 200,
  "status": "200 OK",
  "headers": { },
  "body": { }
}
```

**错误处理：**

| HTTP 状态码 | 错误 kind | 典型 hint |
| --- | --- | --- |
| 401 | `auth` | check whether the token is valid or expired. |
| 403 | `auth` | check whether the token has permission for this API. |
| 其他非 2xx | `api` | check the API response and request parameters. |
| 超时 | `network` | increase --timeout or check network connectivity. |

**示例：**

```bash
zn-cli http GET /api/user/list --query page=1 --query pageSize=20
zn-cli http POST /api/user/create --body '{"name":"test"}'
zn-cli http POST /api/user/create --query page=1 --header X-Test=yes --body '{"name":"test"}'
```

---

### `zn-cli api`

**说明：** 按三层结构渐进式查看 HTTP API 目录：一级大模块 → 二级业务模块 → 三级具体 API。

**当前实现：** 目录数据来自内置 `MockProvider`（`internal/catalog/catalog.go`），**尚未**对接后端动态目录接口。命令执行时不依赖 `token` 与 `base_url`（仅读取配置以确定输出格式）。

**命令：**

```bash
zn-cli api [module] [business] [api] [flags]
```

**参数：**

| 参数 | 说明 |
| --- | --- |
| `[module]` | 一级大模块标识；未传入时返回所有一级大模块 |
| `[business]` | 二级业务模块标识；传入 `[module]` 后可用 |
| `[api]` | 三级 API 标识；传入 `[module]` 和 `[business]` 后可用 |
| `--full` | 返回完整 API 文档数组；仅在传入 `[module]` 和 `[business]`、且未传入 `[api]` 时生效 |

**查询规则：**

| 命令 | 说明 |
| --- | --- |
| `zn-cli api` | 查询所有一级大模块 |
| `zn-cli api <module>` | 查询指定大模块下的二级业务模块 |
| `zn-cli api <module> <business>` | 查询该业务模块下的 API 摘要列表 |
| `zn-cli api <module> <business> --full` | 查询该业务模块下所有 API 的完整详情数组 |
| `zn-cli api <module> <business> <api>` | 查询指定 API 的完整接口文档 |

**参数校验：**

- 最多接受 3 个位置参数
- `--full` 必须与恰好 2 个位置参数一起使用
- 已指定 `[api]` 时不能再使用 `--full`

**Cobra 设计：**

CLI 只静态注册 `api` 一个 Cobra 命令，不把目录数据注册为真实子命令；动态层级通过位置参数解析：

```text
zn-cli
└── api                       # 静态 Cobra 命令
    ├── args[0] = module      # 一级大模块
    ├── args[1] = business    # 二级业务模块
    └── args[2] = api           # 三级 API
```

**动态补全：**

通过 `ValidArgsFunction` 按当前参数层级补全候选项；补全数据同样来自 `MockProvider`，后端不可用时返回空列表。

```text
zn-cli api <TAB>              # 补全一级大模块
zn-cli api ziniao <TAB>       # 补全 ziniao 下的业务模块
zn-cli api ziniao user <TAB>  # 补全 ziniao/user 下的 API
```

**内置目录数据（MockProvider）**

一级大模块：

| name | title |
| --- | --- |
| `ziniao` | 紫鸟业务接口 |
| `erp` | ERP 接口 |

`ziniao` 下业务模块：`user`（用户管理）、`access`（访问策略）、`account`（店铺账号）

`erp` 下业务模块：`order`（订单管理）

`ziniao/user` 下 API：`list`、`detail`、`create`

`ziniao/access` 下 API：`list`

`ziniao/account` 下 API：`list`

**数据模型：**

模块 / 业务模块摘要：

```json
{
  "name": "user",
  "title": "用户管理",
  "description": "用户、员工、权限相关接口"
}
```

API 摘要（`ListAPIs`）额外包含 `method`、`url` 字段。

完整 API 文档（`GetAPI` / `ListFullAPIs`）：

| 字段 | 说明 |
| --- | --- |
| `name` | API 稳定标识 |
| `title` | 展示名称 |
| `description` | 接口说明 |
| `url` | HTTP 路径 |
| `method` | HTTP 方法 |
| `params` | 参数说明（含 `query`、`body` 等） |
| `response` | 响应字段结构说明 |

**输出示例：**

不传参数，返回一级大模块列表（text 模式，列宽 10）：

```bash
zn-cli api
```

```text
ziniao     紫鸟业务接口
erp        ERP 接口
```

传入一级大模块：

```bash
zn-cli api ziniao
```

```text
user       用户管理
access     访问策略
account    店铺账号
```

传入一级大模块和业务模块，返回 API 摘要（含 method、url）：

```bash
zn-cli api ziniao user
```

```text
list       GET    /api/user/list       查询用户列表
detail     GET    /api/user/detail     查询用户详情
create     POST   /api/user/create     创建用户
```

`--full` 返回完整详情数组：

```bash
zn-cli api ziniao user --full
```

```json
[
  {
    "name": "list",
    "title": "查询用户列表",
    "description": "分页查询用户列表",
    "url": "/api/user/list",
    "method": "GET",
    "params": {
      "query": {
        "page": "number",
        "pageSize": "number",
        "keyword": "string"
      },
      "body": null
    },
    "response": {
      "code": "number",
      "message": "string",
      "data": {
        "list": "array",
        "total": "number"
      }
    }
  }
]
```

传入完整三层参数，返回单个 API 文档：

```bash
zn-cli api ziniao user list
```

```json
{
  "name": "list",
  "title": "查询用户列表",
  "description": "分页查询用户列表",
  "url": "/api/user/list",
  "method": "GET",
  "params": {
    "query": {
      "page": "number",
      "pageSize": "number",
      "keyword": "string"
    },
    "body": null
  },
  "response": {
    "code": "number",
    "message": "string",
    "data": {
      "list": "array",
      "total": "number"
    }
  }
}
```

模块、业务模块或 API 不存在时返回 `api` 类错误，hint 为 `run zn-cli api to inspect available modules and APIs.`。标识匹配不区分大小写。

---

### `zn-cli version`

**说明：** 查看当前 CLI 版本信息。

**命令：**

```bash
zn-cli version
```

**参数：** 无（继承全局参数）。

**输出：**

text 模式：

```text
zn-cli 0.1.0
```

json 模式 `data` 字段：

```json
{
  "name": "zn-cli",
  "version": "0.1.0"
}
```

---

## 内部 HTTP 客户端（未暴露为 CLI 命令）

`internal/httpclient` 中另实现了两个占位接口方法，当前**未**注册为 CLI 子命令：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `VerifyAuth` | `GET /auth/verify` | 校验 token |
| `RunRequest` | `GET /request/run` | 执行配置的请求流程 |

如需调用上述接口，请使用 `zn-cli http` 直接请求对应路径。

---

## 常见问题

| 错误信息 | 原因 | 处理建议 |
| --- | --- | --- |
| `token is required` | 未提供 token | 使用 `--token`、`ZINIAO_TOKEN` 或配置文件 |
| `base_url is required` | 未提供基础地址 | 使用 `--base-url`、`ZINIAO_BASE_URL` 或配置文件 |
| `base_url is invalid` | URL 格式不合法 | 使用完整 URL，例如 `https://api.example.com` |
| `path is invalid` | `http` 命令 path 不以 `/` 开头 | 使用 `/api/...` 形式 |
| `body is invalid json` | `--body` 不是合法 JSON | 传入合法 JSON 对象或数组 |
| `invalid key=value pair` | `--query` 或 `--header` 格式错误 | 使用 `key=value` 格式 |
| `output format is invalid` | `--output` 值非法 | 仅支持 `text` 或 `json` |
| `request timed out` | 请求超时 | 增大 `--timeout` 或检查网络 |
| `module/business/api "..." not found` | 目录中无对应项 | 运行 `zn-cli api` 查看可用目录 |
