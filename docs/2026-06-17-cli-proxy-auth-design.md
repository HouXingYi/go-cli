# CLI 代理鉴权与 API 目录设计

**日期**：2026-06-17
**状态**：草稿
**依赖**：[2026-06-11-web-session-auth-design.md](./2026-06-11-web-session-auth-design.md)

---

## 变更记录

| 日期 | 变更内容 |
| --- | --- |
| 2026-06-17 初版 | CLI 代理鉴权体系 + 多服务商 Provider 注册表 + API 目录查询 |

---

## 一、设计背景与目标

### 1.1 背景

`zn-cli` 运行在沙箱环境中，由大模型通过 Function Call 驱动，需要调用后端业务接口（用户管理、店铺查询等）。业务接口由不同服务商（网关 / ERP）提供，各自有不同的鉴权方式。

本服务（`a1-browser-server`）作为沙箱与业务接口之间的**鉴权代理层**：
- 持有用户的凭证（`oauth_string`、`web_session_key` 等），沙箱不接触
- 统一处理加解密、token 刷新
- 提供服务商 API 目录查询，供 CLI 渐进式发现

### 1.2 目标

1. **代理转发**：沙箱通过 HMAC session_token 调用本服务，本服务按服务商的鉴权方式加密/签名后转发
2. **凭证管理**：懒加载 web session 凭证，自动刷新，对调用方透明
3. **API 目录**：后端动态返回可用接口的 OpenAPI 格式目录，三层渐进披露（模块 → 业务 → API）
4. **可扩展**：新增服务商只需注册，不改路由逻辑

---

## 二、整体架构

```
┌─────────────────────────────────────────────────────────┐
│ 沙箱环境                                                  │
│  LLM → Function Call → zn-cli http <method> <path>      │
│                       → zn-cli api [module] [business]   │
│                          └─ HMAC session_token ──┐       │
└──────────────────────────────────────────────────┼───────┘
                                                    │
                                                    ▼
┌──────────────────────────────────────────────────────────┐
│ a1-browser-server (本服务)                                │
│                                                          │
│  /vendor-proxy/cli/{provider}/{path}   ← 代理转发        │
│  /vendor-proxy/cli-api/{provider}      ← API 目录查询    │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │ Provider Registry                                │    │
│  │  ziniao → web_session auth → gateway             │    │
│  │  erp    → app_key auth     → openapi-router      │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │ API Catalog (DB + 本地缓存 + 定时刷新)            │    │
│  │  ziniao/user/{list,detail,create}                │    │
│  │  ziniao/access/{rule/query,...}                  │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────┬───────────────────────────────────┘
                       │ web session / app_key
                       ▼
              ┌─────────────────┐
              │ 目标服务商       │
              │ (网关 / ERP)    │
              └─────────────────┘
```

---

## 三、身份映射链路

```
session_token (HMAC)
  → validate_session_token() → session_id
  → t_enterprise_manage_chat_thread → user_id + company_id
  → Redis cli_web_session:{user_id} → 凭证
```

### 3.1 现有依赖

| 组件 | 说明 |
|------|------|
| `vendor_proxy_token.py` | HMAC session_token 签发/验证，`session_id.expiry_ts.hmac` 格式 |
| `t_enterprise_manage_chat_thread` | 存 `thread_id`(session_id)、`user_id`、`company_id` |
| `VENDOR_PROXY_SECRET` | HMAC 签名密钥，已存在 |

### 3.2 machine_string 策略

| 场景 | machine_string |
|------|---------------|
| 调网关接口 1 | `"sandbox_" + session_id` |
| 网关 Redis Key | `session:{uid}:sandbox_{session_id}` |

---

## 四、数据模型

### 4.1 新增数据库表

#### t_cli_api_module

Provider 一级模块定义，每行一个 service provider。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT PK | 自增主键 |
| `module` | VARCHAR(64) UNIQUE | 一级模块名，如 ziniao, erp |
| `title` | VARCHAR(128) | 展示名称 |
| `description` | VARCHAR(512) | 描述 |
| `status` | TINYINT | 1=正常，0=已废弃/停用 |

#### t_cli_api_endpoint

每个 API 端点一行，按 OpenAPI 结构存储。与 module 一对多关联。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT PK | 自增主键 |
| `module_id` | BIGINT FK | 关联 t_cli_api_module.id |
| `tag` | VARCHAR(64) | OpenAPI tag，即二级 business 分组名 |
| `operation_id` | VARCHAR(128) | OpenAPI operationId，LLM 稳定指代 |
| `summary` | VARCHAR(256) | 接口摘要 |
| `description` | TEXT | 接口功能描述，LLM 判断调用依据 |
| `method` | VARCHAR(8) | GET / POST / PUT / DELETE |
| `path` | VARCHAR(256) | URL 路径，含 {param} 占位符 |
| `status` | TINYINT | 1=正常，0=已废弃/停用 |
| `params` | JSON | OpenAPI parameters 数组及 requestBody |
| `response` | JSON | OpenAPI responses 对象 |
| `example` | JSON | 请求/响应完整示例（x-example） |
| `errors` | JSON | 常见错误码及处理建议（x-errors） |

> 无需 `api_key` 表。身份通过 `session_id → user_id` 解析。

### 4.2 Redis 新增键

```
# oauth_string 缓存
Key:   cli_oauth:{user_id}
Type:  String
TTL:   与 oauth_string 自身有效期对齐（约 15 天）
Value: "<oauth_string>"
写入:  用户登录流程写入（不在本设计范围内）

# web session 凭证（懒加载）
Key:   cli_web_session:{user_id}
Type:  Hash
TTL:   与 refresh_token 有效期对齐
Fields:
  web_session_key   -> Base64 编码的 AES 密钥
  access_token      -> JWT
  refresh_token     -> JWT
  refresh_code      -> 一次性刷新码
  expire_at         -> access_token 过期时间戳（毫秒）
```

### 4.3 凭证获取逻辑

`expire_at` 由网关在接口 1 响应中直接返回，本服务解密后存入 Redis，后续自行比对：

```
首次代理请求：
  GET cli_web_session:{user_id}
  ├─ 命中且 expire_at > now + 60 → 直接使用
  └─ 未命中或 access_token 将过期：
      ├─ GET cli_oauth:{user_id}
      │   └─ 未命中 → 返回 503 "用户未配置鉴权凭证"
      ├─ 调网关 POST /{svc}/web/token/client-auth（接口 1）
      │   响应解密后拿到: web_session_key, access_token, expire_at,
      │                   refresh_token, refresh_code
      │   expire_at = 网关明文返回的过期时间戳（毫秒）
      ├─ 所有字段（含 expire_at）缓存到 cli_web_session:{user_id}
      └─ 返回凭证

access_token 过期（expire_at < now + 60）：
  ├─ 调网关 POST /{svc}/web/token/refresh（接口 2）
  │   响应解密后拿到新: web_session_key, access_token, expire_at, refresh_code
  ├─ 原子更新 cli_web_session:{user_id}
  └─ 重试原请求
```

### 4.4 并发控制

- 凭证首次获取：`Redis SETNX cli_lock:web_session:{user_id}` 分布式锁，TTL 10s
- Token 刷新：同样加锁，并发请求短暂等待或排队

---

## 五、Provider 注册表

### 5.1 数据模型

每个 provider 定义其鉴权方式、目标地址和路径模板：

```python
@dataclass
class ProviderConfig:
    key: str                    # 标识符，如 "ziniao"
    name: str                   # 展示名称，如 "紫鸟业务接口"
    auth_type: Literal["web_session", "app_key"]
    base_url: str               # 目标服务 base URL
    path_template: str          # 路径拼接模板，{rest} 替换为 path 剩余段
    timeout: float = 30.0
    # 仅 web_session
    svc: str | None = None      # 服务名，如 "ent"，用于拼接 /{svc}/web/{rest}
```

### 5.2 预注册 provider

```python
PROVIDERS: dict[str, ProviderConfig] = {
    "ziniao": ProviderConfig(
        key="ziniao",
        name="紫鸟业务接口",
        auth_type="web_session",
        base_url=GATEWAY_BASE_URL,        # 如 https://gateway.ziniao.com
        path_template="/{svc}/web/{rest}",
        svc="ent",
    ),
    "erp": ProviderConfig(
        key="erp",
        name="ERP 接口",
        auth_type="app_key",
        base_url=ERP_API_BASE_URL,        # 如 https://test-sbappstoreapi.ziniao.com/openapi-router
        path_template="/{rest}",
    ),
}
```

### 5.3 鉴权策略

```python
async def resolve_auth(provider: ProviderConfig, user_id: str, company_id: str, session_id: str):
    match provider.auth_type:
        case "web_session":
            return await get_or_refresh_web_session(user_id, session_id)
            # → AuthHeaders(authorization="Bearer <access_token>",
            #               encrypt=WebSessionEncrypt(web_session_key))
            # 请求体需 AES 加密，响应体需 AES 解密
        case "app_key":
            app_key = await fetch_company_app_key(int(company_id))
            # → AuthHeaders(authorization="Bearer <app_key>",
            #               encrypt=None)
            # 请求体/响应体明文透传
```

---

## 六、接口设计

### 6.1 代理转发端点

```
POST /vendor-proxy/cli/{provider}/{path}
```

**鉴权**：`Authorization: Bearer <session_token>`（HMAC）

**Request Body**：

```json
{
  "method": "GET",
  "query": {"page": 1, "pageSize": 20},
  "body": {"keyword": "test"}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `method` | string | 是 | HTTP 方法，转发到目标服务 |
| `query` | object | 否 | URL 查询参数 |
| `body` | object | 否 | 请求体参数 |

**Response Body**（成功）：

```json
{
  "ret": 0,
  "status": "success",
  "msg": "ok",
  "data": { ...业务响应 }
}
```

> `web_session` 模式：本服务解密网关加密响应后，取内层 `ret/status/msg/data` 原样返回。
> `app_key` 模式：目标服务响应直接透传。

**错误响应**：

```json
{
  "ret": 10000,
  "status": "failed",
  "msg": "<原因>",
  "data": null
}
```

**处理流程**：

```
1. HMAC 校验 session_token → session_id
2. 查 t_enterprise_manage_chat_thread → user_id, company_id
3. match provider.auth_type:
   web_session → get_or_refresh_web_session(user_id, session_id)
   app_key     → fetch_company_app_key(company_id)
4. 拼接目标 URL：path_template.format(rest=path, svc=provider.svc)
5. 构造请求：
   web_session → AES 加密 body → {data: "<密文>"}
   app_key     → 明文 body
6. 转发 → 获取响应
7. 处理响应：
   web_session → AES 解密 → 取内层 ret/status/msg/data
   app_key     → 原样返回
8. 返回给沙箱
```

### 6.2 API 目录查询端点

**路由**：

```
GET /vendor-proxy/cli-api[/{module}]
```

**鉴权**：与代理转发相同，`Authorization: Bearer <session_token>`

**三层渐进披露**：URL path 的 `{module}` 对应 CLI 的第一个动态参数，`?business=` 对应第二个，`?api=` 对应第三个。

| CLI 命令 | 请求 | 返回 |
|----------|------|------|
| `zn-cli api` | `GET /vendor-proxy/cli-api` | 一级大模块（provider）列表 |
| `zn-cli api ziniao` | `GET /vendor-proxy/cli-api/ziniao` | ziniao 下业务模块列表 |
| `zn-cli api ziniao user` | `GET /vendor-proxy/cli-api/ziniao?business=user` | user 下 API 摘要列表 |
| `zn-cli api ziniao user --full` | `GET /vendor-proxy/cli-api/ziniao?business=user&full=true` | user 下所有 API 完整详情 |
| `zn-cli api ziniao user list` | `GET /vendor-proxy/cli-api/ziniao?business=user&api=list` | list API 完整文档 |

**查询参数**：

| 参数 | 说明 |
|------|------|
| `business` | 二级业务模块名，与 URL path `{module}` 配合使用 |
| `full` | `true` 时返回业务模块下所有 API 完整详情（仅 business 层级） |
| `api` | 三级 API 名，返回该 API 完整文档 |

**响应结构**：

#### 一级大模块列表（`zn-cli api`）

```
GET /vendor-proxy/cli-api
```

```json
{
  "ret": 0,
  "status": "success",
  "data": {
    "items": [
      {
        "name": "ziniao",
        "title": "紫鸟业务接口",
        "description": "用户、访问策略、店铺账号等业务接口"
      },
      {
        "name": "erp",
        "title": "ERP 接口",
        "description": "企业 ERP 相关接口"
      }
    ]
  }
}
```

#### 业务模块列表（`zn-cli api ziniao`）

```
GET /vendor-proxy/cli-api/ziniao
```

```json
{
  "ret": 0,
  "status": "success",
  "data": {
    "module": "ziniao",
    "items": [
      {
        "name": "user",
        "title": "用户管理",
        "description": "用户、员工、权限相关接口"
      },
      {
        "name": "access",
        "title": "访问策略",
        "description": "访问规则、安全策略相关接口"
      },
      {
        "name": "account",
        "title": "店铺账号",
        "description": "店铺账号管理相关接口"
      }
    ]
  }
}
```

#### API 摘要列表（`zn-cli api ziniao user`）

```
GET /vendor-proxy/cli-api/ziniao?business=user
```

```json
{
  "ret": 0,
  "status": "success",
  "data": {
    "module": "ziniao",
    "business": "user",
    "items": [
      {"name": "list", "title": "查询用户列表", "description": "分页查询用户列表", "method": "GET", "url": "/api/user/list"},
      {"name": "detail", "title": "查询用户详情", "description": "查询单个用户详情", "method": "GET", "url": "/api/user/detail"},
      {"name": "create", "title": "创建用户", "description": "创建新用户", "method": "POST", "url": "/api/user/create"}
    ]
  }
}
```

#### 完整详情 / 具体 API 文档（`zn-cli api ziniao user --full` / `zn-cli api ziniao user list`）

与 CLI 文档对齐，`--full` 返回数组，指定 API 返回单对象：详见 [cli-api-docs.md](./cli-api-docs.md)。

#### 手动刷新缓存

```
POST /vendor-proxy/cli-api/refresh
```

后台管理修改 DB 后调用，立即刷新本地缓存，无需等定时任务触发。

```json
{ "ret": 0, "status": "success", "msg": "ok", "data": {"modules": 2, "refreshed_at": 1749628800} }
```

### 6.3 API 目录数据源

#### 存储架构

```
OpenAPI YAML                        数据库 (持久化)                  本地缓存                CatalogLoader
───────────────────────────────────────────────────────────────────────────────────────────────────────────
外部提供的                           t_cli_api_module                内存 dict               纯内存读取
openapi.yaml  ──导入──→              t_cli_api_endpoint  ──定时刷新──→   {module: [endpoints]}   零延迟
(标准格式)                           (结构化列 + JSON 子列)          (全量，5min)
                                              ↑
                                         后台管理在线修改
```

- **数据库**：两表，按 API 粒度增删改查，支持后台管理表单操作
- **本地内存缓存**：定时全量加载 → 组装 OpenAPI 格式 → 注入 LLM context
- **导入能力**：标准 OpenAPI 3.0 YAML/JSON 直接导入到两表中

#### 三层映射关系

```
CLI 层级    数据来源
──────────────────────────────────────────────────────
1 module    t_cli_api_module.module 列
2 business  t_cli_api_endpoint.tag 列（WHERE module_id=? GROUP BY tag）
3 api       t_cli_api_endpoint.operation_id 列（WHERE tag=?）
```

#### 数据库表

```
t_cli_api_module:
┌────┬──────────┬──────────────────┬────────────────────────────────┬────────┐
│ id │ module   │ title            │ description                    │ status │
├────┼──────────┼──────────────────┼────────────────────────────────┼────────┤
│  1 │ ziniao   │ 紫鸟业务接口     │ 用户、访问策略、店铺账号等     │      1 │
│  2 │ erp      │ ERP 接口         │ 企业 ERP 相关接口              │      1 │
└────┴──────────┴──────────────────┴────────────────────────────────┴────────┘

t_cli_api_endpoint:
┌────┬───────────┬──────┬──────────────┬───────────────┬────────┬─────────────────────┬────────┬────────┬────────┐
│ id │ module_id │ tag  │ operation_id │ summary       │ method │ path                │ status │ params │ ...    │
├────┼───────────┼──────┼──────────────┼───────────────┼────────┼─────────────────────┼────────┼────────┼────────┤
│  1 │         1 │ user │ listUsers    │ 查询用户列表  │ GET    │ /api/user/list      │      1 │ {...}  │        │
│  2 │         1 │ user │ getUser      │ 查询用户详情  │ GET    │ /api/user/{id}      │      1 │ {...}  │        │
│  3 │         1 │ user │ createUser   │ 创建用户      │ POST   │ /api/user/create    │      1 │ {...}  │        │
│  4 │         1 │ access│ listRules   │ 查询访问规则  │ GET    │ /api/access/rule    │      1 │ {...}  │        │
└────┴───────────┴──────┴──────────────┴───────────────┴────────┴─────────────────────┴────────┴────────┴────────┘
```

#### 结构列与 JSON 列划分

| 列 | 类型 | 为什么 |
|----|------|--------|
| `tag`, `method`, `path`, `status` | **结构化列** | 需要 WHERE / GROUP BY / ORDER BY |
| `params`, `response`, `example`, `errors` | **JSON 列** | 复杂嵌套，LLM 直接消费，不需要 SQL 条件查询 |

#### OpenAPI 导入

支持从标准 OpenAPI 3.0 YAML/JSON 直接导入到两表中：

```
OpenAPI 字段                              → 表列
──────────────────────────────────────────────────────────────
info.title                                → t_cli_api_module.title
info.description                          → t_cli_api_module.description
paths[path][method].tags[0]               → t_cli_api_endpoint.tag
paths[path][method].operationId           → t_cli_api_endpoint.operation_id
paths[path][method].summary               → t_cli_api_endpoint.summary
paths[path][method].description           → t_cli_api_endpoint.description
method (GET/POST/...)                      → t_cli_api_endpoint.method
path                                       → t_cli_api_endpoint.path
paths[path][method].parameters            → t_cli_api_endpoint.params (JSON)
paths[path][method].responses             → t_cli_api_endpoint.response (JSON)
paths[path][method].x-example             → t_cli_api_endpoint.example (JSON)
paths[path][method].x-errors              → t_cli_api_endpoint.errors (JSON)
```

**导入接口**：

```
POST /vendor-proxy/cli-api/import
Content-Type: multipart/form-data
Body: file=<openapi.yaml>, module=<module_name>
```

或管理后台直接上传。同一个 module 重复导入时全量替换（先删旧 endpoint 再插入新数据）。

**导入实现**：

```python
async def import_openapi(module_name: str, spec: dict):
    # 1) upsert module
    module = await upsert_module(
        module=module_name,
        title=spec["info"]["title"],
        description=spec["info"].get("description", ""),
    )

    # 2) 删旧 endpoint
    await delete_endpoints_by_module(module.id)

    # 3) 遍历 paths，每个 method 一行
    for path, methods in spec.get("paths", {}).items():
        for method in ("get", "post", "put", "delete", "patch"):
            op = methods.get(method)
            if not op:
                continue

            await insert_endpoint(
                module_id=module.id,
                tag=op.get("tags", [None])[0],
                operation_id=op.get("operationId", ""),
                summary=op.get("summary", ""),
                description=op.get("description", ""),
                method=method.upper(),
                path=path,
                params=json.dumps(_build_params(path, methods, op)),
                response=json.dumps(op.get("responses", {})),
                example=json.dumps(op.get("x-example", {})),
                errors=json.dumps(op.get("x-errors", [])),
            )
```

#### CatalogLoader 实现

```python
class CatalogLoader:
    """按层级查询 API 目录，本地内存缓存 + 定时 DB 刷新."""

    _cache: dict[str, list[EndpointRow]] = {}  # module → endpoints

    async def refresh(self) -> None:
        """定时任务：JOIN 两表 → 组装 OpenAPI 格式 → 原子替换 _cache."""

    async def list_modules(self) -> list[ModuleSummary]:
        """一级：遍历 _cache.keys() → 返回 provider 列表."""

    async def list_businesses(self, module: str) -> list[BusinessSummary]:
        """二级：_cache[module] 按 tag GROUP BY → 返回 tag 列表."""

    async def list_apis(self, module: str, business: str) -> list[ApiSummary]:
        """三级摘要：_cache[module] WHERE tag=? → 返回摘要."""

    async def get_api(self, module: str, business: str, api: str) -> ApiSpec | None:
        """三级详情：_cache[module] WHERE operation_id=? → 返回完整 OpenAPI 定义."""

    async def list_apis_full(self, module: str, business: str) -> list[ApiSpec]:
        """--full：_cache[module] WHERE tag=? → 返回完整详情数组."""
```

#### 缓存刷新机制

```
启动时     → refresh() JOIN 两表 → 组装 OpenAPI → 写入 _cache
每 5 分钟  → 定时任务 refresh()
手动刷新  → POST /vendor-proxy/cli-api/refresh
导入后    → 自动触 refresh()
```

#### 注入 LLM 的方式

`refresh()` 时将数据库行组装为标准 OpenAPI 格式，`list_apis_full()` 直接返回完整 OpenAPI paths 对象，LLM 原生理解。

#### 文件结构

```
src/core/api_catalog/
├── __init__.py
├── catalog_loader.py    # 本地内存缓存 + 定时 DB 刷新 + OpenAPI 组装
├── importer.py          # OpenAPI YAML/JSON 导入（spec → 两表）
├── models.py            # ModuleRow, EndpointRow, ModuleSummary, ... ApiSpec
├── crud.py              # t_cli_api_module / t_cli_api_endpoint CRUD
└── router.py            # 查询 + 导入 + 刷新 + 后台管理 CRUD
```

---

## 七、安全设计

### 7.1 鉴权边界

| 边界 | 鉴权方式 | 密钥持有方 |
|------|---------|-----------|
| 沙箱 → 本服务 | HMAC session_token | 沙箱持有 session_token，本服务持有 VENDOR_PROXY_SECRET |
| 本服务 → web session 网关 | JWT access_token + AES 加密 | 仅本服务持有 web_session_key / access_token |
| 本服务 → ERP | app_key Bearer token | 仅本服务持有（查表） |

### 7.2 凭证隔离

- `oauth_string`：仅存于 Redis，沙箱 / LLM 不可达
- `web_session_key` / `access_token` / `refresh_token`：仅存于 Redis + 本服务内存
- `app_key`：仅存于 DB + 本服务内存缓存（60s TTL）
- 沙箱只持有短期 HMAC token

### 7.3 Token 自动刷新

**触发方式**：被动、按需。每次代理请求进入时检查，无后台定时任务。

```
每次代理请求 → HGET cli_web_session:{user_id} expire_at
  ├─ expire_at > now + 60 → 无需刷新，直接使用
  └─ expire_at <= now + 60 → 需要刷新：
      1. 加锁 cli_lock:web_session:{user_id}（SETNX, TTL 10s）
      2. 读 refresh_token + refresh_code
      3. 调网关 POST /{svc}/web/token/refresh
      4. 响应用旧 web_session_key 解密
      5. 用网关返回的新 expire_at 更新 Redis
      6. 释放锁 → 用新凭证转发
```

**expire_at 来源**：
- 接口 1（client-auth）响应明文 `data.data.expire_at`（毫秒时间戳），由网关返回
- 接口 2（token/refresh）响应同样包含新的 `expire_at`
- 本服务解密后写入 Redis，后续自行读取比对，不依赖网关

**失败处理**：
- 刷新失败 → 返回 502 "会话刷新失败，请重新登录"
- refresh_token 过期（同 oauth TTL）→ 需用户重新登录以刷新 cli_oauth:{user_id}

### 7.4 错误安全

| 场景 | HTTP 状态码 | ret | 说明 |
|------|-----------|-----|------|
| session_token 无效/过期 | 401 | — | HMAC 校验失败 |
| session_id 未绑定 user_id | 401 | — | 表无记录 |
| oauth_string 无缓存 | 503 | — | 用户未配置鉴权凭证 |
| 网关接口 1 调用失败 | 502 | — | 换取 Web 会话失败 |
| access_token 刷新失败 | 502 | — | 会话刷新失败 |
| 网关业务错误 | 200 | 透传 | 解密后返回网关原始错误码 |
| ERP 业务错误 | 200 | 透传 | 原样返回 |

---

## 八、AES 加解密实现要点

### 8.1 算法

沿用文档 §2.2 的 AES/CBC/NoPadding 算法：

```
加密/解密密钥 = web_session_key（Base64 解码后的原始字节）
```

| 角色 | 操作 | key 范围 | IV 范围 |
|------|------|---------|--------|
| 本服务 | 加密请求 → 网关 | `[0, 16)` | `[len-16, len)` |
| 本服务 | 解密网关响应 | `[1, 17)` | `[len-17, len-1)` |

### 8.2 依赖

- `pycryptodome`（已有，用于 HuBu RSA）
- 需确认 `Crypto.Cipher.AES` 支持 CBC/NoPadding

### 8.3 实现模块

新增 `src/util/web_session_crypto.py`：
- `encrypt_request(plain: dict, web_session_key: bytes) -> str` — 返回 Base64 密文
- `decrypt_response(data: str, web_session_key: bytes) -> dict` — 返回解密后的 JSON
- `_get_key_iv(ws_key: bytes, mode: str) -> tuple[bytes, bytes]` — 按规范截取 key/IV

### 8.4 Padding

网关使用 NoPadding，需调用方手动 PKCS7 padding：

```python
def _pkcs7_pad(data: bytes, block_size: int = 16) -> bytes:
    pad_len = block_size - len(data) % block_size
    return data + bytes([pad_len] * pad_len)

def _pkcs7_unpad(data: bytes) -> bytes:
    pad_len = data[-1]
    return data[:-pad_len]
```

---

## 九、错误码定义

新增错误码段：`cli_proxy`

| 错误码 | 说明 |
|--------|------|
| `30001` | 会话未绑定用户 |
| `30002` | 用户未配置鉴权凭证（oauth_string 无缓存） |
| `30003` | Web 会话获取失败 |
| `30004` | Web 会话刷新失败 |
| `30005` | Provider 不存在 |
| `30006` | API 目录未找到 |
| `30007` | 网关请求失败 |

---

## 十、配置项汇总

```python
# ── CLI 代理配置 ──
# 网关地址（web session auth）
CLI_PROXY_GATEWAY_BASE_URL: str       # 如 https://gateway.ziniao.com
CLI_PROXY_GATEWAY_SVC: str = "ent"   # 默认服务名

# ERP 地址（app_key auth，已有）
ENTERPRISE_MANAGE_AGENT_API_BASE_URL: str
ERP_API_BASE_URL: str

# HMAC（已有，复用）
VENDOR_PROXY_SECRET: str

# API 目录
CLI_API_CATALOG_REFRESH_INTERVAL: int = 300  # 本地缓存定时刷新间隔（秒）

# 并发控制
CLI_WEB_SESSION_LOCK_TTL: int = 10    # 凭证获取分布式锁 TTL（秒）

# Token 提前刷新阈值
CLI_TOKEN_REFRESH_BEFORE_SECONDS: int = 60  # 提前 60 秒刷新
```

---

## 十一、文件结构规划

```
src/
├── core/
│   ├── api_catalog/                    # 新增：API 目录
│   │   ├── __init__.py
│   │   ├── catalog_loader.py           # 本地内存缓存 + 定时刷新 + OpenAPI 组装
│   │   ├── importer.py                 # OpenAPI YAML/JSON 导入（spec → 两表）
│   │   ├── models.py                   # ModuleRow, EndpointRow, ModuleSummary, ... ApiSpec
│   │   ├── crud.py                     # t_cli_api_module / t_cli_api_endpoint CRUD
│   │   └── router.py                   # 查询 + 导入 + 刷新 + 后台管理 CRUD
│   ├── config.py                       # 追加 CLI 代理配置
│   └── middleware/
│       └── jwt_auth_middleware.py       # 追加 /vendor-proxy/cli 白名单
├── model/
│   ├── cli_api_module.py               # 新增：t_cli_api_module ORM
│   └── cli_api_endpoint.py             # 新增：t_cli_api_endpoint ORM
├── cli_proxy/                          # 新增：CLI 代理模块
│   ├── __init__.py
│   ├── router.py                       # 路由注册（/cli/ + /cli-api/）
│   ├── provider_registry.py            # Provider 注册表
│   ├── auth/
│   │   ├── __init__.py
│   │   ├── base.py                     # AuthStrategy 抽象
│   │   ├── web_session.py              # web_session auth 策略
│   │   └── app_key.py                  # app_key auth 策略
│   ├── service.py                      # 核心代理服务（身份解析、请求构造、凭证管理）
│   └── exceptions.py                   # CLI 代理异常
└── util/
    └── web_session_crypto.py           # 新增：AES 加解密工具
```
