# CLI 代理接口 — 外部调用指南

**服务器地址**：`https://agent-swarm-dev.ziniao.com`  
**最后更新**：2026-06-24

---

## 1. 获取 token

联系服务管理员提供 `session_token`。

测试环境可用 token（有效期到 2026-06-25）：

```
e4461536-b229-44c6-8225-23c3e9a64e05.1782371920.324b32c728617050efa9fa91144099fea5f6eefec80516b96ee8e19b990dc8e4
```

---

## 2. 接口

所有接口的 base 路径：

```
https://agent-swarm-dev.ziniao.com/api/v1/claw/cli-proxy
```

### 2.1 API 目录查询

| 说明 | 路径 |
|------|------|
| 所有模块列表 | `GET /cli-api` |
| 某模块的 business 列表 | `GET /cli-api/{module}` |
| 某 business 下所有 API | `GET /cli-api/{module}/apis?business={biz}` |
| 单个 API 完整定义 | `GET /cli-api/{module}/apis?business={biz}&api={name}` |

### 2.2 代理转发

**路径**：`POST /cli/{module}/{gateway-path}`

**请求体**（JSON）：

```json
{
  "method": "<网关接口的HTTP方法>",
  "body": {<网关接口入参>}
}
```

`{gateway-path}` 就是目录接口返回的 `url` 原样拼接上去。比如目录返回 `/ent/web/v5/simple/company/info`，转发路径就是 `/cli/ziniao/ent/web/v5/simple/company/info`。

当前支持模块：`ziniao`、`erp`

---

## 3. curl 示例

```bash
TOKEN="<你的token>"
BASE="https://agent-swarm-dev.ziniao.com/api/v1/claw/cli-proxy"

# --- API 目录 ---

# 所有模块
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/cli-api" | python3 -m json.tool

# company business 的接口列表
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/cli-api/ziniao/apis?business=company" | python3 -m json.tool

# 单个接口详情
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/cli-api/ziniao/apis?business=company&api=getCompanyInfo" | python3 -m json.tool

# --- 代理转发 ---

# 查询公司信息 (body {} 即可)
curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"method":"POST","body":{}}' \
  "$BASE/cli/ziniao/ent/web/v5/simple/company/info" | python3 -m json.tool

# 查询用户列表 (分页)
curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"method":"POST","body":{"page_num":1,"page_size":5}}' \
  "$BASE/cli/ziniao/ent/web/v5/simple/user/list" | python3 -m json.tool

# 查询店铺列表
curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"method":"POST","body":{"page_num":1,"page_size":3}}' \
  "$BASE/cli/ziniao/store/web/v5/simple/account/list" | python3 -m json.tool

# 查询登陆日志
curl -s -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"method":"POST","body":{"start_date":"2026-06-01","end_date":"2026-06-24","page_num":1,"page_size":10}}' \
  "$BASE/cli/ziniao/log/web/v5/login-log/list" | python3 -m json.tool
```

---

## 4. 常见错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 连接失败 | 网络不通 | 确认能访问 `agent-swarm-dev.ziniao.com` |
| `401` | token 过期或无效 | 联系管理员获取新 token |
| `30005` | 模块名不在支持列表中 | 当前仅支持 `ziniao`、`erp` |
| `30006` | API 目录查询的 module/business/api 名不正确 | 先用 `/cli-api` 查看可用列表 |
| `MethodNotAllowed` | proxy body 中的 `method` 写错 | 对照 API 目录确认正确方法（GET/POST） |
