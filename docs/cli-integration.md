# CLI 对接文档

## 一、环境变量

沙箱启动时，`a1-browser-server` 自动注入以下环境变量：

| 环境变量 | 说明 | 示例值 |
|----------|------|--------|
| `VENDOR_PROXY_TOKEN` | HMAC session_token，`session_id.expiry_ts.hmac` 格式 | `abc123.1749632400.def456...` |
| `VENDOR_PROXY_BASE` | 代理 URL 前缀 | `https://api.example.com/api/v1/claw/vendor-proxy` |

`VENDOR_PROXY_TOKEN` 随沙箱生命周期绑定，沙箱销毁后 token 同时失效。

---

## 二、鉴权

所有 CLI → 后端请求统一使用 HMAC session_token：

```http
Authorization: ${VENDOR_PROXY_TOKEN}
```

Token 格式：`{session_id}.{expiry_unix_seconds}.{hmac_sha256_hex}`

**错误响应**（token 无效）：
```json
HTTP 401
{"detail": "Invalid or expired session token"}
```

---

## 三、代理转发

### 3.1 请求

```http
POST ${VENDOR_PROXY_BASE}/cli/{provider}/{path}
Authorization: ${VENDOR_PROXY_TOKEN}
Content-Type: application/json
```

**路径参数**：

| 参数 | 说明 | 可选值 |
|------|------|--------|
| `provider` | 服务商标识 | `ziniao` (紫鸟业务)、`erp` (ERP) |
| `path` | 业务接口路径 | 如 `user/list` |

**请求体**：

```json
{
  "method": "GET",
  "query": {"page": 1, "pageSize": 20},
  "body": {"keyword": "test"}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `method` | string | 否 | HTTP 方法（默认 GET） |
| `query` | object | 否 | URL 查询参数 |
| `body` | object | 否 | 请求体参数 |

### 3.2 成功响应

```json
{
  "ret": 0,
  "status": "success",
  "msg": "ok",
  "data": { ... }
}
```

### 3.3 错误响应

| HTTP | ret | 说明 |
|------|-----|------|
| 401 | — | session_token 无效或过期 |
| 404 | — | provider 不存在 |
| 200 | 10000 | 上游网关认证失败 |
| 200 | 30001 | 会话未绑定用户 |
| 200 | 30002 | 用户未配置鉴权凭证 |
| 200 | 30003 | Web 会话获取失败 |
| 200 | 30007 | 网关请求失败 |

---

## 四、API 目录查询

### 4.1 一级：模块列表

```http
GET ${VENDOR_PROXY_BASE}/cli-api
Authorization: ${VENDOR_PROXY_TOKEN}
```

响应：
```json
{
  "code": 200, "error_code": 0, "message": "success",
  "data": {
    "items": [
      {"name": "ziniao", "title": "紫鸟业务接口", "description": "..."},
      {"name": "erp", "title": "ERP 接口", "description": "..."}
    ]
  }
}
```

### 4.2 二级：业务模块

```http
GET ${VENDOR_PROXY_BASE}/cli-api/{module}
Authorization: ${VENDOR_PROXY_TOKEN}
```

### 4.3 三级：API 摘要

```http
GET ${VENDOR_PROXY_BASE}/cli-api/{module}/apis?business={tag}
Authorization: ${VENDOR_PROXY_TOKEN}
```

### 4.4 三级详情（--full）

```http
GET ${VENDOR_PROXY_BASE}/cli-api/{module}/apis?business={tag}&full=true
Authorization: ${VENDOR_PROXY_TOKEN}
```

### 4.5 单个 API 文档

```http
GET ${VENDOR_PROXY_BASE}/cli-api/{module}/apis?business={tag}&api={operation_id}
Authorization: ${VENDOR_PROXY_TOKEN}
```

---

## 五、CLI 命令映射

| CLI 命令 | HTTP 请求 |
|----------|----------|
| `zn-cli api` | `GET /vendor-proxy/cli-api` |
| `zn-cli api ziniao` | `GET /vendor-proxy/cli-api/ziniao` |
| `zn-cli api ziniao user` | `GET /vendor-proxy/cli-api/ziniao/apis?business=user` |
| `zn-cli api ziniao user --full` | `GET /vendor-proxy/cli-api/ziniao/apis?business=user&full=true` |
| `zn-cli http GET /user/list` | `POST /vendor-proxy/cli/ziniao/user/list` |

---

## 六、LLM Function Call 示例

```json
{
  "name": "zn_cli_http",
  "description": "调用后端业务接口。",
  "parameters": {
    "type": "object",
    "properties": {
      "provider": {"type": "string", "enum": ["ziniao", "erp"]},
      "method": {"type": "string", "enum": ["GET", "POST", "PUT", "DELETE"]},
      "path": {"type": "string"},
      "query": {"type": "object"},
      "body": {"type": "object"}
    },
    "required": ["provider", "method", "path"]
  }
}
```
