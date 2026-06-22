# zn-cli 文档

## CLI 名称

`zn-cli`（`internal/config.AppName`，当前版本 `0.1.0`）

可执行文件入口：`cmd/ziniao/main.go`，构建产物通常命名为 `ziniao` 或 `ziniao.exe`。

## 命令总览

```text
zn-cli [command] [flags]

Commands:
  auth      Configure CLI authentication
  http      Send an authenticated HTTP request via vendor-proxy
  api       Inspect backend HTTP API catalog
  version   Print CLI version
```

不带子命令运行 `zn-cli` 时显示帮助信息。

## 环境变量

沙箱环境由 `a1-browser-server` 注入以下变量；本地开发未设置 `VENDOR_PROXY_BASE` 时，CLI 使用内置 MockBackend 跑通命令。

| 环境变量 | 说明 |
| --- | --- |
| `CLI_AUTH_KEY` | 鉴权密钥，请求以 `Authorization: Bearer <key>` 发送 |
| `VENDOR_PROXY_BASE` | vendor-proxy 基础 URL，例如 `https://api.example.com/api/v1/claw/vendor-proxy` |

设置 `VENDOR_PROXY_BASE` 后，CLI 自动切换为 HTTPBackend 对接真实后端；未设置时 `http` 与 `api` 均走 MockBackend。

以下值为内置常量（`internal/config/config.go`），当前不可通过 CLI 或环境变量修改：

| 常量 | 当前值 | 说明 |
| --- | --- | --- |
| `DefaultTimeout` | `10s` | HTTP 请求超时 |

## 输出格式

当前 CLI 固定使用 text 输出。成功时向 stdout 输出人类可读消息；失败时向 stderr 输出错误与提示：

```text
Error: auth key is required
Hint: set CLI_AUTH_KEY.
```

`internal/output` 包保留 json 输出能力，后续可通过 Viper 扩展 `--output` 等配置项。

---

## 核心命令

### `zn-cli auth`

**说明：** 校验 `CLI_AUTH_KEY` 环境变量是否已设置。当前实现**不会**向后端发送请求，也**不会**将配置写入磁盘。

**命令：**

```bash
zn-cli auth
```

**参数：** 无。

**前置条件：**

- `CLI_AUTH_KEY` 非空

**输出：**

```text
Authentication configured successfully.
```

**示例：**

```bash
export CLI_AUTH_KEY=your-key
zn-cli auth
```

---

### `zn-cli http`

**说明：** 通过 vendor-proxy 代理转发业务请求。CLI 将用户指定的 HTTP 方法、路径、查询参数和请求体包装为代理请求，发往 `POST {VENDOR_PROXY_BASE}/cli/{provider}/{path}`。

**命令：**

```bash
zn-cli http <method> <path> --provider <provider> [flags]
```

**参数：**

| 参数 | 说明 |
| --- | --- |
| `<method>` | 业务 HTTP 方法，例如 `GET`、`POST`、`PUT`、`DELETE`（大小写不敏感） |
| `<path>` | 业务接口路径，须以 `/` 开头，例如 `/api/user/list`；path 以 catalog 返回的 `url` 为准 |

**命令级 flags：**

| 参数 | 说明 |
| --- | --- |
| `--provider` | 服务商标识，**必填**，例如 `ziniao`、`erp` |
| `--query <key=value>` | 追加 URL 查询参数，可重复传入 |
| `--body <json>` | 请求体，须为合法 JSON，适用于 `POST`、`PUT` 等 |

**请求行为：**

- 实际 HTTP 方法固定为 `POST`（代理协议）
- 目标 URL：`{VENDOR_PROXY_BASE}/cli/{provider}/{trimmedPath}`
- 代理请求体：`{"method":"GET","query":{...},"body":{...}}`
- 自动设置 `Authorization: Bearer <CLI_AUTH_KEY>`
- 成功时输出响应信封中的 `data` 字段（JSON 美化）

**前置条件：** 需要设置 `CLI_AUTH_KEY`。

**错误处理：**

| 场景 | 错误 kind | 典型 hint |
| --- | --- | --- |
| 未设置 `CLI_AUTH_KEY` | `config` | set CLI_AUTH_KEY. |
| 未传 `--provider` | `config` | use --provider ziniao or --provider erp. |
| HTTP 401 | `auth` | check whether CLI_AUTH_KEY is valid. |
| HTTP 200 且 `ret != 0` | `api` | 按 ret 码提示（如 30002 用户未配置鉴权凭证） |
| 超时 | `network` | check network connectivity. |

**示例：**

```bash
export CLI_AUTH_KEY=your-key
export VENDOR_PROXY_BASE=https://api.example.com/api/v1/claw/vendor-proxy

zn-cli http GET /api/user/list --provider ziniao --query page=1 --query pageSize=20
zn-cli http POST /api/user/create --provider ziniao --body '{"name":"test"}'
```

本地无后端时（未设置 `VENDOR_PROXY_BASE`），MockBackend 根据 catalog 中的 path/method 返回模拟 `data`。

---

### `zn-cli api`

**说明：** 按三层结构渐进式查看 HTTP API 目录：一级大模块 → 二级业务模块 → 三级具体 API。与 `http` 共用同一 `Backend` 抽象。

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

| 命令 | 后端请求（HTTPBackend） |
| --- | --- |
| `zn-cli api` | `GET /cli-api` |
| `zn-cli api <module>` | `GET /cli-api/{module}` |
| `zn-cli api <module> <business>` | `GET /cli-api/{module}?business={business}` |
| `zn-cli api <module> <business> --full` | `GET /cli-api/{module}?business={business}&full=true` |
| `zn-cli api <module> <business> <api>` | `GET /cli-api/{module}?business={business}&api={api}` |

**参数校验：**

- 最多接受 3 个位置参数
- `--full` 必须与恰好 2 个位置参数一起使用
- 已指定 `[api]` 时不能再使用 `--full`

**Cobra 设计：**

```text
zn-cli
└── api                       # 静态 Cobra 命令
    ├── args[0] = module      # 一级大模块
    ├── args[1] = business    # 二级业务模块
    └── args[2] = api         # 三级 API
```

**动态补全：**

通过 `ValidArgsFunction` 按当前参数层级补全候选项；数据来自当前 Backend（Mock 或远程）。

**MockBackend 目录数据**

一级大模块：

| name | title |
| --- | --- |
| `ziniao` | 紫鸟业务接口 |
| `erp` | ERP 接口 |

`ziniao` 下业务模块：`user`（用户管理）、`access`（访问策略）、`account`（店铺账号）

`erp` 下业务模块：`order`（订单管理）

`ziniao/user` 下 API：`list`、`detail`、`create`

**数据模型：**

模块 / 业务模块摘要：

```json
{
  "name": "user",
  "title": "用户管理",
  "description": "用户、员工、权限相关接口"
}
```

API 摘要额外包含 `method`、`url` 字段。

完整 API 文档（`--full` / 指定 `[api]`）：

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

```bash
zn-cli api
```

```text
ziniao     紫鸟业务接口
erp        ERP 接口
```

```bash
zn-cli api ziniao user
```

```text
list       GET    /api/user/list       查询用户列表
detail     GET    /api/user/detail     查询用户详情
create     POST   /api/user/create     创建用户
```

模块、业务模块或 API 不存在时返回 `api` 类错误，hint 为 `run zn-cli api to inspect available modules and APIs.`。标识匹配不区分大小写。

---

### `zn-cli version`

**说明：** 查看当前 CLI 版本信息。

**命令：**

```bash
zn-cli version
```

**输出：**

```text
zn-cli 0.1.0
```

---

## 常见问题

| 错误信息 | 原因 | 处理建议 |
| --- | --- | --- |
| `auth key is required` | 未设置 `CLI_AUTH_KEY` | 执行 `export CLI_AUTH_KEY=...` |
| `provider is required` | `http` 未传 `--provider` | 添加 `--provider ziniao` 或 `--provider erp` |
| `body is invalid json` | `--body` 不是合法 JSON | 传入合法 JSON 对象或数组 |
| `invalid key=value pair` | `--query` 格式错误 | 使用 `key=value` 格式 |
| `request timed out` | 请求超时 | 检查网络连接 |
| `module/business/api "..." not found` | 目录中无对应项 | 运行 `zn-cli api` 查看可用目录 |
