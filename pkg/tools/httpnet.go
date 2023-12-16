package tools

type HttpResponse struct {
	Message string      `json:"message",omitempty`
	Data    interface{} `json:"data",omitempty`
	Total   interface{} `json:"total",omitempty`
}
