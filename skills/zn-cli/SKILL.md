---
name: zn-cli
description: Patterns for invoking zn-cli from agents. Covers auth, module defaults, api discovery, http proxy calls, and error recovery.
---

# 参考

## 非交互策略

`zn-cli` 没有交互式提示。在非 TTY 环境（Agent shell、管道）中：

- 缺少必填输入会立即失败，stderr 输出 `Error:`，并附带一行 `Hint:` 说明修复方式。
- 没有 pager；内容直接写入 stdout。
- 不要臆造 `--token`、`--interactive` 等不存在的 flag。

应从 stderr 解析失败信息，不要只看退出码：

```text
Error: module is required
Hint: run zn-cli config module set <name> or pass --module.
```

## 推荐工作流

任务开始时按以下顺序执行：

1. `zn-cli agent guide` — 加载本文档（若已在上下文中可跳过）。
2. `zn-cli auth` — 确认 `CLI_AUTH_KEY` 已设置（不发起网络请求）。
3. `zn-cli api` — 列出一级大模块（如 `ziniao`、`erp`）。
4. `zn-cli config module set <module>` — 持久化 module，后续 `http` 无需每次传 `--module`。
5. `zn-cli api <module> <business>` — 列出 API；用 `zn-cli api <module> <business> <api>` 查看参数文档。
6. `zn-cli http <method> <path> [--query ...] [--body ...]` — 调用 API。

仅在临时切换大模块时使用 `http --module <name>`；否则依赖已持久化的默认值。

## 鉴权

- 在调用 `http` 或远程 `api` 之前，于环境中设置 `CLI_AUTH_KEY`。
- `zn-cli auth` 仅检查该变量非空；不联系后端，也不将密钥写入磁盘。
- 切勿在命令行传入密钥。日志与错误输出不会回显密钥。

## Module 定位

`http` 按以下优先级解析 module（与 `api [module]`、后端 `{provider}` 同义）：

1. `--module` flag（最高）
2. `ZINIAO_MODULE` 环境变量
3. `zn-cli config module set` 持久化在 `state.yaml` 中的值

若均未设置，`http` 报错 `module is required`。

```bash
zn-cli config module set ziniao
zn-cli config module get
zn-cli http GET /api/user/list --module erp   # 临时覆盖
```

## API 发现（`zn-cli api`）

渐进式目录查询 — 不要死记硬背 API 列表，运行时动态发现：

```bash
zn-cli api                              # 一级大模块
zn-cli api <module>                     # 二级业务模块
zn-cli api <module> <business>          # API 摘要（method、url）
zn-cli api <module> <business> --full   # 该业务下全部 API 完整文档
zn-cli api <module> <business> <api>    # 单个 API 文档（参数、响应）
```

规则：

- 位置参数最多 3 个。
- `--full` 必须与 `[module]` 和 `[business]` 一起使用；不能与 `[api]` 组合。
- 标识匹配不区分大小写。
- 调用 `http` 时使用目录返回的 `url` 和 `method`；path 必须以 `/` 开头。

## HTTP 调用（`zn-cli http`）

```bash
zn-cli http <method> <path> [--module <name>] [--query '<json-object>'] [--body '<json>']
```

- `<method>`：`GET`、`POST`、`PUT`、`DELETE`（大小写不敏感）。
- `<path>`：目录中的业务路径 `url`，例如 `/api/user/list`。
- `--query`：用于 URL 查询参数的 JSON **对象**。
- `--body`：任意合法 JSON（对象或数组）作为请求体。

CLI 会将调用包装为 vendor-proxy 的 POST；你只需指定逻辑上的 HTTP 方法和路径。

示例：

```bash
zn-cli http GET /api/user/list --query '{"page":1,"pageSize":20}'
zn-cli http POST /api/user/create --body '{"name":"test"}'
```

在 shell 中用单引号包裹 JSON，避免转义问题。

## 输出解析

- **成功：** stdout 输出人类可读文本或 JSON 块。`http` 将响应 `data` 字段以缩进 JSON 打印；`api` 查询单个 API 时输出 JSON，列表视图输出制表符对齐文本。
- **失败：** stderr 输出 `Error:` 和 `Hint:`；退出码非零。
- 优先根据 `Hint:` 决定下一步，不要盲目重试。

## Mock 与远程后端

| 条件 | 行为 |
| --- | --- |
| 未设置 `VENDOR_PROXY_BASE` | MockBackend — `api` 和 `http` 离线可用，使用内置示例目录 |
| 已设置 `VENDOR_PROXY_BASE` | HTTPBackend — 真实 vendor-proxy；需要有效的 `CLI_AUTH_KEY` |

沙箱环境会自动注入 `VENDOR_PROXY_BASE` 和 `CLI_AUTH_KEY`。

## 安全

- `CLI_AUTH_KEY` 仅通过环境变量提供；没有 `--token` 或密钥配置文件。
- 不要在 Agent 消息中记录或重复密钥。
- 状态文件（`state.yaml`）只保存默认 module 名称，不保存凭证。

## 常见错误

| 错误 | 原因 | 恢复方式 |
| --- | --- | --- |
| `auth key is required` | 未设置 `CLI_AUTH_KEY` | 设置环境变量；运行 `zn-cli auth` 验证 |
| `module is required` | 无默认 module 且未传 `--module` | `zn-cli config module set <name>` 或传入 `--module` |
| `body is invalid json` | `--body` 不是合法 JSON | 使用单引号包裹的 JSON 字符串 |
| `query is invalid json` | `--query` 不是 JSON 对象 | 例如 `'{"page":1}'` |
| `request timed out` | 网络超时（默认 10s） | 检查网络 / 代理 |
| `module/business/api "..." not found` | 目录中无对应项 | 运行 `zn-cli api` 查看有效名称 |
| HTTP 401 / 鉴权错误 | 密钥无效 | 与运维确认 `CLI_AUTH_KEY` |

## 业务域 Skill

本 Skill 仅说明**如何驱动 zn-cli**。业务工作流（账号、员工、设备、访问策略等）优先查阅对应的 `ziniao-*` 业务 Skill。调用 `http` 前，务必用 `zn-cli api <module> <business> <api>` 确认参数结构。
