package catalog

import (
	"context"
	"fmt"
	"strings"

	"ziniao/internal/apperr"
)

type ModuleSummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type BusinessSummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type APISummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Method      string `json:"method"`
	URL         string `json:"url"`
}

type APIDocument struct {
	Name        string      `json:"name"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	URL         string      `json:"url"`
	Method      string      `json:"method"`
	Params      interface{} `json:"params"`
	Response    interface{} `json:"response"`
}

type Provider interface {
	ListModules(ctx context.Context) ([]ModuleSummary, error)
	ListBusinesses(ctx context.Context, module string) ([]BusinessSummary, error)
	ListAPIs(ctx context.Context, module, business string) ([]APISummary, error)
	ListFullAPIs(ctx context.Context, module, business string) ([]APIDocument, error)
	GetAPI(ctx context.Context, module, business, api string) (APIDocument, error)
}

type MockProvider struct {
	modules    []ModuleSummary
	businesses map[string][]BusinessSummary
	apis       map[string]map[string][]APIDocument
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		modules: []ModuleSummary{
			{Name: "ziniao", Title: "紫鸟业务接口", Description: "用户、访问策略、店铺账号等业务接口"},
			{Name: "erp", Title: "ERP 接口", Description: "企业 ERP 相关接口"},
		},
		businesses: map[string][]BusinessSummary{
			"ziniao": {
				{Name: "user", Title: "用户管理", Description: "用户、员工、权限相关接口"},
				{Name: "access", Title: "访问策略", Description: "访问规则、安全策略相关接口"},
				{Name: "account", Title: "店铺账号", Description: "店铺账号管理相关接口"},
			},
			"erp": {
				{Name: "order", Title: "订单管理", Description: "ERP 订单相关接口"},
			},
		},
		apis: map[string]map[string][]APIDocument{
			"ziniao": {
				"user": {
					{
						Name:        "list",
						Title:       "查询用户列表",
						Description: "分页查询用户列表",
						URL:         "/api/user/list",
						Method:      "GET",
						Params: map[string]interface{}{
							"query": map[string]string{
								"page":     "number",
								"pageSize": "number",
								"keyword":  "string",
							},
							"body": nil,
						},
						Response: map[string]interface{}{
							"code":    "number",
							"message": "string",
							"data": map[string]string{
								"list":  "array",
								"total": "number",
							},
						},
					},
					{
						Name:        "detail",
						Title:       "查询用户详情",
						Description: "查询单个用户详情",
						URL:         "/api/user/detail",
						Method:      "GET",
						Params: map[string]interface{}{
							"query": map[string]string{
								"id": "string",
							},
							"body": nil,
						},
						Response: map[string]interface{}{
							"code":    "number",
							"message": "string",
							"data":    "object",
						},
					},
					{
						Name:        "create",
						Title:       "创建用户",
						Description: "创建新用户",
						URL:         "/api/user/create",
						Method:      "POST",
						Params: map[string]interface{}{
							"query": map[string]string{},
							"body": map[string]string{
								"name": "string",
							},
						},
						Response: map[string]interface{}{
							"code":    "number",
							"message": "string",
							"data":    "object",
						},
					},
				},
				"access": {
					{
						Name:        "list",
						Title:       "查询访问规则",
						Description: "分页查询访问策略规则",
						URL:         "/api/access/rule",
						Method:      "GET",
						Params: map[string]interface{}{
							"query": map[string]string{
								"page":     "number",
								"pageSize": "number",
							},
							"body": nil,
						},
						Response: map[string]interface{}{
							"code":    "number",
							"message": "string",
							"data":    "object",
						},
					},
				},
				"account": {
					{
						Name:        "list",
						Title:       "查询店铺账号",
						Description: "分页查询店铺账号列表",
						URL:         "/api/account/list",
						Method:      "GET",
						Params: map[string]interface{}{
							"query": map[string]string{
								"page":     "number",
								"pageSize": "number",
							},
							"body": nil,
						},
						Response: map[string]interface{}{
							"code":    "number",
							"message": "string",
							"data":    "object",
						},
					},
				},
			},
		},
	}
}

func (p *MockProvider) ListModules(ctx context.Context) ([]ModuleSummary, error) {
	return append([]ModuleSummary(nil), p.modules...), nil
}

func (p *MockProvider) ListBusinesses(ctx context.Context, module string) ([]BusinessSummary, error) {
	module = normalize(module)
	if _, err := p.findModule(module); err != nil {
		return nil, err
	}
	items := p.businesses[module]
	return append([]BusinessSummary(nil), items...), nil
}

func (p *MockProvider) ListAPIs(ctx context.Context, module, business string) ([]APISummary, error) {
	docs, err := p.listDocs(module, business)
	if err != nil {
		return nil, err
	}

	items := make([]APISummary, 0, len(docs))
	for _, doc := range docs {
		items = append(items, APISummary{
			Name:        doc.Name,
			Title:       doc.Title,
			Description: doc.Description,
			Method:      doc.Method,
			URL:         doc.URL,
		})
	}
	return items, nil
}

func (p *MockProvider) ListFullAPIs(ctx context.Context, module, business string) ([]APIDocument, error) {
	docs, err := p.listDocs(module, business)
	if err != nil {
		return nil, err
	}
	return append([]APIDocument(nil), docs...), nil
}

func (p *MockProvider) GetAPI(ctx context.Context, module, business, api string) (APIDocument, error) {
	api = normalize(api)
	docs, err := p.listDocs(module, business)
	if err != nil {
		return APIDocument{}, err
	}
	for _, doc := range docs {
		if doc.Name == api {
			return doc, nil
		}
	}
	return APIDocument{}, notFound("api", api)
}

func (p *MockProvider) listDocs(module, business string) ([]APIDocument, error) {
	module = normalize(module)
	business = normalize(business)
	if _, err := p.findModule(module); err != nil {
		return nil, err
	}
	if !p.hasBusiness(module, business) {
		return nil, notFound("business", business)
	}
	docs := p.apis[module][business]
	return append([]APIDocument(nil), docs...), nil
}

func (p *MockProvider) findModule(module string) (ModuleSummary, error) {
	for _, item := range p.modules {
		if item.Name == module {
			return item, nil
		}
	}
	return ModuleSummary{}, notFound("module", module)
}

func (p *MockProvider) hasBusiness(module, business string) bool {
	for _, item := range p.businesses[module] {
		if item.Name == business {
			return true
		}
	}
	return false
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func notFound(kind, name string) error {
	return apperr.New(apperr.KindAPI, fmt.Sprintf("%s %q not found", kind, name), "run zn-cli api to inspect available modules and APIs.")
}
