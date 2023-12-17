package tools

type HttpResponse struct {
	Message         string      `json:"message",omitempty`
	Data            interface{} `json:"data",omitempty`
	RecordsTotal    interface{} `json:"recordsTotal",omitempty`
	RecordsFiltered interface{} `json:"recordsFiltered",omitempty`
}
