

# zn-cli 文档

## cli名字
zn-cli

多cli（暂定）
zn-manage-cli
zn-operate-cli


## 核心接口

### `zn-cli auth`

**说明：** 配置 CLI 鉴权信息。
**命令：**
```bash
zn-cli auth
```
**参数：**
暂无。
**输出：**
鉴权成功后，CLI 会保存认证配置，并提示配置成功。


### `zn-cli http`
**说明：** 向后端发送带鉴权信息的 HTTP 请求，并打印接口响应。
**命令：**
```bash
zn-cli http <method> <path> [options]
```
**参数：**

| 参数 | 说明 |
| --- | --- |
| `<method>` | HTTP 请求方法，例如 `GET`、`POST`、`PUT`、`DELETE` |
| `<path>` | 后端接口路径，例如 `/api/user/list` |
| `--query <key=value>` | 追加 URL 查询参数，可重复传入 |
| `--body <json>` | 请求体内容，适用于 `POST`、`PUT` 等请求 |
| `--header <key=value>` | 追加请求头，可重复传入 |

**输出：**

打印后端接口返回的响应内容。默认输出 JSON；如果请求失败，输出错误码、错误信息和排查建议。

示例：

```bash
zn-cli http GET /api/user/list --query page=1 --query pageSize=20
zn-cli http POST /api/user/create --body '{"name":"test"}'
```

### `zn-cli api`

**说明：** 查询后端返回的 HTTP API 目录，并按三层结构渐进式披露：一级大模块、二级业务模块、三级具体 API。模块、业务模块和 API 均不在 CLI 内置写死，而是在命令执行时向后端动态查询。

**命令：**
```bash
zn-cli api [module] [business] [api] [flags]
```

**参数：**

| 参数 | 说明 |
| --- | --- |
| `[module]` | 一级大模块标识。未传入时，返回所有一级大模块 |
| `[business]` | 二级业务模块标识。传入 `[module]` 后可用，返回该业务模块下的 API 列表 |
| `[api]` | 三级 API 标识。传入 `[module]` 和 `[business]` 后可用，返回具体 API 文档 |
| `--full` | 返回完整 API 文档。仅在传入 `[module]` 和 `[business]`、未传入 `[api]` 时生效 |

**查询规则：**

| 命令 | 说明 |
| --- | --- |
| `zn-cli api` | 查询所有一级大模块，例如 `ziniao`、`provider` |
| `zn-cli api <module>` | 查询指定一级大模块下的二级业务模块，例如 `user`、`access`、`account` |
| `zn-cli api <module> <business>` | 查询指定业务模块下的 API 摘要列表 |
| `zn-cli api <module> <business> --full` | 查询指定业务模块下所有 API 的完整详情数组 |
| `zn-cli api <module> <business> <api>` | 查询指定 API 的完整接口文档 |

**Cobra 设计：**

CLI 只静态注册 `api` 一个 Cobra 命令，不把后端动态返回的模块、业务模块、API 注册成真实的 Cobra 子命令。

原因如下：

1. Cobra 需要在执行命令逻辑前先完成命令树解析。如果动态子命令没有提前注册，用户输入的动态模块名会被判定为 unknown command。
2. 如果在启动时提前请求后端并注册完整命令树，会导致 CLI 启动依赖网络和鉴权状态，`--help`、补全、错误提示都会受到后端可用性的影响。
3. 动态 API 目录本质是业务数据，更适合作为参数解析，而不是 CLI 程序结构的一部分。

推荐结构：

```text
zn-cli
└── api                       # 静态 Cobra 命令
    ├── args[0] = module      # 动态一级大模块
    ├── args[1] = business    # 动态二级业务模块
    └── args[2] = api         # 动态三级 API
```

`api` 命令根据参数数量决定查询层级：

```text
0 个参数 -> 查询一级大模块列表
1 个参数 -> 查询该大模块下的业务模块列表
2 个参数 -> 查询该业务模块下的 API 摘要列表
2 个参数 + --full -> 查询该业务模块下所有 API 的完整详情数组
3 个参数 -> 查询具体 API 文档
```

`--full` 是业务模块层的输出模式，不属于 API 标识。若已经传入 `[api]`，则不能再使用 `--full`。

**动态补全：**

虽然不把动态数据注册成真实子命令，但可以通过 Cobra 的 `ValidArgsFunction` 实现动态补全：

```text
zn-cli api <TAB>
# 动态补全一级大模块

zn-cli api ziniao <TAB>
# 动态补全 ziniao 下的业务模块

zn-cli api ziniao user <TAB>
# 动态补全 ziniao/user 下的 API
```

补全逻辑同样通过后端接口获取候选项；如果后端不可用，补全可以降级为空列表或提示错误，但不影响基础命令结构。

**后端目录接口建议：**

后端应提供 API 目录查询能力，CLI 按层级请求：

| 场景 | 请求含义 |
| --- | --- |
| 未传参数 | 查询一级大模块 |
| 传入 `module` | 查询该大模块下的业务模块 |
| 传入 `module` + `business` | 查询该业务模块下的 API 摘要列表 |
| 传入 `module` + `business` + `--full` | 查询该业务模块下所有 API 的完整详情数组 |
| 传入 `module` + `business` + `api` | 查询具体 API 文档 |

后端返回的数据应包含稳定标识和展示名称，例如：

```json
{
  "items": [
    {
      "name": "user",
      "title": "用户管理",
      "description": "用户、员工、权限相关接口"
    }
  ]
}
```

具体 API 文档应包含以下信息：

| 字段 | 说明 |
| --- | --- |
| `name` | API 稳定标识，用于命令参数 |
| `title` | API 展示名称 |
| `description` | HTTP 接口说明 |
| `url` | HTTP 接口地址 |
| `method` | HTTP 请求方法，例如 `GET`、`POST`、`PUT`、`DELETE` |
| `params` | 参数格式，包含 query、path、body 等参数说明 |
| `response` | 返回格式，说明接口响应字段结构 |

**输出：**

根据传入参数数量返回不同层级的数据。

不传参数时，返回一级大模块列表：

```bash
zn-cli api
```

```text
ziniao    紫鸟业务接口
provider  服务商接口
```

传入一级大模块时，返回业务模块列表：

```bash
zn-cli api ziniao
```

```text
user      用户管理
access    访问策略
account   店铺账号
```

传入一级大模块和业务模块时，返回 API 列表：

```bash
zn-cli api ziniao user
```

```text
list      查询用户列表
detail    查询用户详情
create    创建用户
```

传入一级大模块、业务模块和 `--full` 时，返回该业务模块下所有 API 的完整详情数组：

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
  },
  {
    "name": "detail",
    "title": "查询用户详情",
    "description": "查询单个用户详情",
    "url": "/api/user/detail",
    "method": "GET",
    "params": {
      "query": {
        "id": "string"
      },
      "body": null
    },
    "response": {
      "code": "number",
      "message": "string",
      "data": "object"
    }
  }
]
```

传入完整三层参数时，返回具体 API 文档：

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

因此，当前建议采用“静态 `api` 命令 + 动态参数 + 动态补全”的设计。


## 其他
### `zn-cli version`

**说明：** 查看当前 CLI 的版本信息。
**命令：**
```bash
zn-cli version
```
**参数：**
暂无。
**输出：**
输出当前 CLI 的版本号，用于确认本地安装的 CLI 版本。



