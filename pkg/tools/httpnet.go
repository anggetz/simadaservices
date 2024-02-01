package tools

type HttpResponse struct {
	Message         string      `json:"message",omitempty`
	Data            interface{} `json:"data",omitempty`
	RecordsTotal    interface{} `json:"recordsTotal",omitempty`
	RecordsFiltered interface{} `json:"recordsFiltered",omitempty`
}

type HttpResponseReport struct {
	Message         string      `json:"message"`
	Data            interface{} `json:"data"`
	RecordsTotal    interface{} `json:"recordsTotal"`
	RecordsFiltered interface{} `json:"recordsFiltered"`
	Draw            int64       `json:"draw"`
	SummaryPerPage  interface{} `json:"summary_perpage"`
	SummaryPage     interface{} `json:"summary_page"`
}

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type NotOkResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
