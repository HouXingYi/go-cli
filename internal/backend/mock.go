package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ziniao/internal/apperr"
)

type mockModule struct {
	Name        string
	Title       string
	Description string
}

type mockBusiness struct {
	Name        string
	Title       string
	Description string
}

type mockAPI struct {
	Name        string
	Title       string
	Description string
	URL         string
	Method      string
	Params      interface{}
	RequestBody interface{}
	Example     interface{}
	Response    interface{}
}

type MockBackend struct {
	modules    []mockModule
	businesses map[string][]mockBusiness
	apis       map[string]map[string][]mockAPI
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		modules: []mockModule{
			{Name: "ziniao", Title: "紫鸟业务接口", Description: "用户、访问策略、店铺账号等业务接口"},
			{Name: "erp", Title: "ERP 接口", Description: "企业 ERP 相关接口"},
		},
		businesses: map[string][]mockBusiness{
			"ziniao": {
				{Name: "user", Title: "用户管理", Description: "用户、员工、权限相关接口"},
				{Name: "access", Title: "访问策略", Description: "访问规则、安全策略相关接口"},
				{Name: "account", Title: "店铺账号", Description: "店铺账号管理相关接口"},
			},
			"erp": {
				{Name: "order", Title: "订单管理", Description: "ERP 订单相关接口"},
			},
		},
		apis: map[string]map[string][]mockAPI{
			"ziniao": {
				"user": {
					{
						Name: "list", Title: "查询用户列表", Description: "分页查询用户列表",
						URL: "/api/user/list", Method: "GET",
						Params: map[string]interface{}{
							"query": map[string]string{"page": "number", "pageSize": "number", "keyword": "string"},
							"body":  nil,
						},
						RequestBody: map[string]interface{}{
							"required": false,
						},
						Example: map[string]interface{}{
							"request": map[string]int{"page": 1},
						},
						Response: map[string]interface{}{
							"code": "number", "message": "string",
							"data": map[string]string{"list": "array", "total": "number"},
						},
					},
					{
						Name: "detail", Title: "查询用户详情", Description: "查询单个用户详情",
						URL: "/api/user/detail", Method: "GET",
						Params: map[string]interface{}{
							"query": map[string]string{"id": "string"},
							"body":  nil,
						},
						Response: map[string]interface{}{
							"code": "number", "message": "string", "data": "object",
						},
					},
					{
						Name: "create", Title: "创建用户", Description: "创建新用户",
						URL: "/api/user/create", Method: "POST",
						Params: map[string]interface{}{
							"query": map[string]string{},
							"body":  map[string]string{"name": "string"},
						},
						Response: map[string]interface{}{
							"code": "number", "message": "string", "data": "object",
						},
					},
				},
				"access": {
					{
						Name: "list", Title: "查询访问规则", Description: "分页查询访问策略规则",
						URL: "/api/access/rule", Method: "GET",
						Params: map[string]interface{}{
							"query": map[string]string{"page": "number", "pageSize": "number"},
							"body":  nil,
						},
						Response: map[string]interface{}{
							"code": "number", "message": "string", "data": "object",
						},
					},
				},
				"account": {
					{
						Name: "list", Title: "查询店铺账号", Description: "分页查询店铺账号列表",
						URL: "/api/account/list", Method: "GET",
						Params: map[string]interface{}{
							"query": map[string]string{"page": "number", "pageSize": "number"},
							"body":  nil,
						},
						Response: map[string]interface{}{
							"code": "number", "message": "string", "data": "object",
						},
					},
				},
			},
		},
	}
}

func (b *MockBackend) Proxy(ctx context.Context, module string, method, path string, query, body json.RawMessage) (json.RawMessage, error) {
	module = normalize(module)
	method = strings.ToUpper(strings.TrimSpace(method))
	trimmed := trimPath(path)

	for business, docs := range b.apis[module] {
		_ = business
		for _, doc := range docs {
			if !pathMatches(doc.URL, trimmed) {
				continue
			}
			if doc.Method != method {
				return nil, apperr.New(apperr.KindAPI, fmt.Sprintf("method %s does not match API %s", method, doc.Name), "check method and path against zn-cli api output.")
			}
			return json.Marshal(mockProxyResponse(doc, query, body))
		}
	}

	return nil, apperr.New(apperr.KindAPI, fmt.Sprintf("path %q not found for module %q", path, module), "run zn-cli api to inspect available modules and APIs.")
}

func (b *MockBackend) Catalog(ctx context.Context, module, business, api string, full bool) (json.RawMessage, error) {
	module = normalize(module)
	business = normalize(business)
	api = normalize(api)

	switch {
	case module == "":
		items := make([]map[string]string, 0, len(b.modules))
		for _, item := range b.modules {
			items = append(items, map[string]string{
				"name": item.Name, "title": item.Title, "description": item.Description,
			})
		}
		return mustMarshal(map[string]interface{}{"items": items})
	case business == "":
		if !b.hasModule(module) {
			return nil, notFound("module", module)
		}
		items := make([]map[string]string, 0, len(b.businesses[module]))
		for _, item := range b.businesses[module] {
			items = append(items, map[string]string{
				"name": item.Name, "title": item.Title, "description": item.Description,
			})
		}
		return mustMarshal(map[string]interface{}{"module": module, "items": items})
	case api == "" && full:
		docs, err := b.listDocs(module, business)
		if err != nil {
			return nil, err
		}
		return mustMarshal(docsToFull(docs))
	case api == "":
		docs, err := b.listDocs(module, business)
		if err != nil {
			return nil, err
		}
		items := make([]map[string]string, 0, len(docs))
		for _, doc := range docs {
			items = append(items, map[string]string{
				"name": doc.Name, "title": doc.Title, "description": doc.Description,
				"method": doc.Method, "url": doc.URL,
			})
		}
		return mustMarshal(map[string]interface{}{"module": module, "business": business, "items": items})
	default:
		docs, err := b.listDocs(module, business)
		if err != nil {
			return nil, err
		}
		for _, doc := range docs {
			if doc.Name == api {
				return mustMarshal(docToFull(doc))
			}
		}
		return nil, notFound("api", api)
	}
}

func (b *MockBackend) ListModuleNames() []string {
	names := make([]string, 0, len(b.modules))
	for _, item := range b.modules {
		names = append(names, item.Name)
	}
	return names
}

func (b *MockBackend) ListBusinessNames(module string) []string {
	module = normalize(module)
	items := b.businesses[module]
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}

func (b *MockBackend) ListAPINames(module, business string) []string {
	docs, err := b.listDocs(normalize(module), normalize(business))
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(docs))
	for _, doc := range docs {
		names = append(names, doc.Name)
	}
	return names
}

func (b *MockBackend) hasModule(module string) bool {
	for _, item := range b.modules {
		if item.Name == module {
			return true
		}
	}
	return false
}

func (b *MockBackend) listDocs(module, business string) ([]mockAPI, error) {
	if !b.hasModule(module) {
		return nil, notFound("module", module)
	}
	if !b.hasBusiness(module, business) {
		return nil, notFound("business", business)
	}
	docs := b.apis[module][business]
	return append([]mockAPI(nil), docs...), nil
}

func (b *MockBackend) hasBusiness(module, business string) bool {
	for _, item := range b.businesses[module] {
		if item.Name == business {
			return true
		}
	}
	return false
}

func pathMatches(apiURL, requestPath string) bool {
	apiPath := trimPath(apiURL)
	if apiPath == requestPath {
		return true
	}
	if strings.TrimPrefix(apiPath, "api/") == requestPath {
		return true
	}
	return false
}

func mockProxyResponse(doc mockAPI, query, body json.RawMessage) map[string]interface{} {
	result := map[string]interface{}{
		"mock":   true,
		"api":    doc.Name,
		"method": doc.Method,
		"url":    doc.URL,
	}
	if queryMap := rawMessageToMap(query); len(queryMap) > 0 {
		result["query"] = queryMap
	}
	if bodyMap := rawMessageToMap(body); len(bodyMap) > 0 {
		result["body"] = bodyMap
	}
	return result
}

func docsToFull(docs []mockAPI) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		items = append(items, docToFull(doc))
	}
	return items
}

func docToFull(doc mockAPI) map[string]interface{} {
	out := map[string]interface{}{
		"name": doc.Name, "title": doc.Title, "description": doc.Description,
		"url": doc.URL, "method": doc.Method, "params": doc.Params, "response": doc.Response,
	}
	if doc.RequestBody != nil {
		out["requestBody"] = doc.RequestBody
	}
	if doc.Example != nil {
		out["example"] = doc.Example
	}
	return out
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func notFound(kind, name string) error {
	return apperr.New(apperr.KindAPI, fmt.Sprintf("%s %q not found", kind, name), "run zn-cli api to inspect available modules and APIs.")
}

func mustMarshal(value interface{}) (json.RawMessage, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindAPI, "failed to encode mock response", "check catalog data.", err)
	}
	return payload, nil
}
