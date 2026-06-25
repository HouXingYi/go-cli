# 紫鸟 CLI 命令参考

适用于 `zn-ent`、`zn-eco`（及未来同架构变体）。命令、flags、环境变量与输出格式完全一致。差异仅在于：二进制名、`UserConfigDir()` 下状态子目录，以及 HTTP 请求头 `Cli-Type`（编译期内置，非环境变量）。下文示例统一使用 `zn-ent`；使用 `zn-eco` 时将命令名替换即可。

## 变体与构建

| 产物 | 入口 | 内置 cli-type |
| --- | --- | --- |
| `zn-ent` | `cmd/zn-ent/main.go` | `zn-ent` |
| `zn-eco` | `cmd/zn-eco/main.go` | `zn-eco` |

变体注册表：`internal/variant`。`AppName`、状态目录、`Cli-Type` 均由编译变体决定，当前版本 `0.1.0`。

## 命令总览

```text
<cli> [command] [flags]    # <cli> = zn-ent | zn-eco

Commands:
  agent     Agent-oriented CLI guidance
  auth      Configure CLI authentication
  config    Manage CLI configuration
  http      Send an authenticated HTTP request via vendor-proxy
  api       Inspect backend HTTP API catalog
  version   Print CLI version
```

不带子命令运行时显示帮助信息。

## 环境变量

沙箱环境由 `a1-browser-server` 注入以下变量；本地开发未设置 `CLI_PROXY_BASE` 时，CLI 使用内置 MockBackend 跑通命令。

| 环境变量 | 说明 |
| --- | --- |
| `CLI_PROXY_TOKEN` | 鉴权密钥，请求以 `Authorization: Bearer <key>` 发送 |
| `CLI_PROXY_BASE` | vendor-proxy 基础 URL，例如 `https://api.example.com/api/v1/claw/vendor-proxy` |
| `ZINIAO_MODULE` | 默认一级大模块（module）；优先级低于 `http --module`，高于持久化配置 |
| `ZINIAO_CONFIG_DIR` | CLI 状态目录；未设置时使用 `UserConfigDir()/<AppName>/`（`zn-ent` 或 `zn-eco`） |

设置 `CLI_PROXY_BASE` 后，CLI 自动切换为 HTTPBackend 对接真实后端；未设置时 `http` 与 `api` 均走 MockBackend。远程请求自动携带 `Cli-Type` 头（值因变体而异）。

**术语说明：** CLI 统一使用 **module** 表示一级大模块（与 `api [module]` 一致）。后端对接文档中的 URL 路径段 `{provider}` 与 module 指同一值，见 [cli-integration.md](cli-integration.md)。

以下值为内置常量（`internal/config/config.go`），当前不可通过 CLI 或环境变量修改：

| 常量 | 当前值 | 说明 |
| --- | --- | --- |
| `DefaultTimeout` | `10s` | HTTP 请求超时 |

## 输出格式

当前 CLI 固定使用 text 输出。成功时向 stdout 输出人类可读消息；失败时向 stderr 输出错误与提示：

```text
Error: auth key is required
Hint: set CLI_PROXY_TOKEN.
```

`internal/output` 包保留 json 输出能力，后续可通过 Viper 扩展 `--output` 等配置项。

---

## 核心命令

### agent guide

**说明：** 向 stdout 输出内置 Agent 使用指南（模板 `skills/shared/SKILL.md`，按变体渲染 `AppName`）。无需 `CLI_PROXY_TOKEN`，不访问网络。输出为原始 Markdown，不经 `Renderer` 包装，供 Agent 自举加载上下文。

**命令：**

```bash
zn-ent agent guide
```

**参数：** 无。

**输出：** 完整 SKILL 正文（含 YAML frontmatter）。

---

### auth

**说明：** 校验 `CLI_PROXY_TOKEN` 环境变量是否已设置。当前实现**不会**向后端发送请求，也**不会**将配置写入磁盘。

**命令：**

```bash
zn-ent auth
```

**参数：** 无。

**前置条件：**

- `CLI_PROXY_TOKEN` 非空

**输出：**

```text
Authentication configured successfully.
```

**示例：**

```bash
export CLI_PROXY_TOKEN=your-key
zn-ent auth
```

---

### config module

**说明：** 管理默认一级大模块（module）。`http` 在未传 `--module` 时会使用已配置的默认值，避免 LLM 在每次请求中重复传递 module。

**子命令：**

```bash
zn-ent config module set <module>   # 持久化默认 module
zn-ent config module get            # 查看当前默认 module
zn-ent config module clear          # 清除持久化的默认 module
```

**持久化位置：** `$ZINIAO_CONFIG_DIR/state.yaml`（或 `UserConfigDir()/zn-ent/state.yaml` 等），内容示例：

```yaml
module: ziniao
```

**示例：**

```bash
zn-ent api
zn-ent config module set ziniao
zn-ent config module get
# ziniao

zn-ent config module clear
```

**说明：** `config module get` 在设置了 `ZINIAO_MODULE` 时优先显示环境变量中的值（`source: environment`）；否则显示持久化配置（`source: config`）。

---

### http

**说明：** 通过 vendor-proxy 代理转发业务请求。CLI 将用户指定的 HTTP 方法、路径、查询参数和请求体包装为代理请求，发往 `POST {CLI_PROXY_BASE}/cli/{module}/{path}`（后端文档中该路径段有时写作 `{provider}`，与 module 同义）。

**命令：**

```bash
zn-ent http <method> <path> [flags]
```

**参数：**

| 参数 | 说明 |
| --- | --- |
| `<method>` | 业务 HTTP 方法，例如 `GET`、`POST`、`PUT`、`DELETE`（大小写不敏感） |
| `<path>` | 业务接口路径，须以 `/` 开头，例如 `/api/user/list`；path 以 catalog 返回的 `url` 为准 |

**命令级 flags：**

| 参数 | 说明 |
| --- | --- |
| `--module` | 一级大模块标识，**可选**；传入时临时覆盖默认 module，例如 `ziniao`、`erp` |
| `--query <json>` | URL 查询参数，须为合法 JSON 对象 |
| `--body <json>` | 请求体，须为合法 JSON，适用于 `POST`、`PUT` 等 |

**Module 解析优先级（高 → 低）：**

1. `--module` flag
2. `ZINIAO_MODULE` 环境变量
3. `config module set` 持久化的值（`state.yaml`）

均未设置时报错，hint 引导执行 `zn-ent config module set <name>` 或 `zn-ent api` 查看可选 module。

**请求行为：**

- 实际 HTTP 方法固定为 `POST`（代理协议）
- 目标 URL：`{CLI_PROXY_BASE}/cli/{module}/{trimmedPath}`
- 代理请求体：`{"method":"GET","query":{...},"body":{...}}`
- 自动设置 `Authorization: Bearer <CLI_PROXY_TOKEN>`
- 自动设置 `Cli-Type: <编译变体值>`
- 成功时输出响应信封中的 `data` 字段（JSON 美化）

**前置条件：** 需要设置 `CLI_PROXY_TOKEN`；需要已配置默认 module 或传入 `--module`。

**错误处理：**

| 场景 | 错误 kind | 典型 hint |
| --- | --- | --- |
| 未设置 `CLI_PROXY_TOKEN` | `config` | set CLI_PROXY_TOKEN. |
| 未配置 module | `config` | run zn-ent config module set \<name\> or pass --module. |
| HTTP 401 | `auth` | check whether CLI_PROXY_TOKEN is valid. |
| HTTP 200 且 `ret != 0` | `api` | 按 ret 码提示（如 30002 用户未配置鉴权凭证） |
| 超时 | `network` | check network connectivity. |

**推荐工作流（LLM 沙箱）：**

```bash
export CLI_PROXY_TOKEN=your-key
export CLI_PROXY_BASE=https://api.example.com/api/v1/claw/vendor-proxy

zn-ent api
zn-ent config module set ziniao
zn-ent http GET /api/user/list --query '{"page":1,"pageSize":20}'
zn-ent http POST /api/user/create --body '{"name":"test"}'

# 临时切换大模块
zn-ent http GET /api/order/list --module erp
```

本地无后端时（未设置 `CLI_PROXY_BASE`），MockBackend 根据 catalog 中的 path/method 返回模拟 `data`。

---

### api

**说明：** 按三层结构渐进式查看 HTTP API 目录：一级大模块 → 二级业务模块 → 三级具体 API。与 `http` 共用同一 `Backend` 抽象。

**命令：**

```bash
zn-ent api [module] [business] [api] [flags]
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
| `zn-ent api` | `GET /cli-api` |
| `zn-ent api <module>` | `GET /cli-api/{module}` |
| `zn-ent api <module> <business>` | `GET /cli-api/{module}/apis?business={business}` |
| `zn-ent api <module> <business> --full` | `GET /cli-api/{module}/apis?business={business}&full=true` |
| `zn-ent api <module> <business> <api>` | `GET /cli-api/{module}/apis?business={business}&api={api}` |

**参数校验：**

- 最多接受 3 个位置参数
- `--full` 必须与恰好 2 个位置参数一起使用
- 已指定 `[api]` 时不能再使用 `--full`

**Cobra 设计：**

```text
<cli>
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
zn-ent api
```

```text
ziniao     紫鸟业务接口
erp        ERP 接口
```

```bash
zn-ent api ziniao user
```

```text
list       GET    /api/user/list       查询用户列表
detail     GET    /api/user/detail     查询用户详情
create     POST   /api/user/create     创建用户
```

模块、业务模块或 API 不存在时返回 `api` 类错误，hint 为 `run zn-ent api to inspect available modules and APIs.`。标识匹配不区分大小写。

---

### version

**说明：** 查看当前 CLI 版本与变体信息。

**命令：**

```bash
zn-ent version
```

**输出：**

```text
zn-ent 0.1.0 (cli-type: zn-ent)
```

---

## LLM 调用说明

沙箱内 LLM 推荐按以下顺序调用：

1. `zn-ent agent guide` — 加载 Agent 使用规范（可选，若已在上下文中可跳过）
2. `zn-ent auth` — 确认 `CLI_PROXY_TOKEN` 已配置
3. `zn-ent api` — 列出可用大模块
4. `zn-ent config module set <module>` — 持久化当前任务的大模块
5. `zn-ent http <method> <path> [--query ...] [--body ...]` — 发送业务请求（无需每次传 module）

临时切换大模块时使用 `http --module`。`module` 与后端对接文档 URL 路径中的 `{provider}` 同义。业务 API 参数以 `zn-ent api <module> <business> <api>` 为准，勿在上下文中硬编码 API 列表。

---

## 常见问题

| 错误信息 | 原因 | 处理建议 |
| --- | --- | --- |
| `auth key is required` | 未设置 `CLI_PROXY_TOKEN` | 执行 `export CLI_PROXY_TOKEN=...` |
| `module is required` | `http` 未配置默认 module 且未传 `--module` | 执行 `zn-ent config module set ziniao` 或添加 `--module ziniao` |
| `body is invalid json` | `--body` 不是合法 JSON | 传入合法 JSON 对象或数组 |
| `query is invalid json` | `--query` 不是合法 JSON 对象 | 传入合法 JSON 对象，例如 `'{"page":1}'` |
| `request timed out` | 请求超时 | 检查网络连接 |
| `module/business/api "..." not found` | 目录中无对应项 | 运行 `zn-ent api` 查看可用目录 |
