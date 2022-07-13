package elastic

import "time"

type doc struct {
	Attack    string    `json:"attack"`
	Uuid      string    `json:"uuid"`
	Code      int       `json:"code"`
	Timestamp time.Time `json:"timestamp"`
	Latency   int       `json:"latency"`
	BytesOut  int       `json:"bytes_out"`
	BytesIn   int       `json:"bytes_in"`
	Error     string    `json:"error"`
	Body      string    `json:"body"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	HasError  bool      `json:"has_error"`
	HasBody   bool      `json:"has_body"`
	Version   string    `json:"version"`
	Headers   Headers	
}

type Headers struct {
	Content_Length []string  `json:"Content-Length"`
	Content_Type []string    `json:"Content-Type"`
	Date []string            `json:"Date"`
	Server []string          `json:"Server"`
	Vary []string            `json:"Vary"`
	X_Envoy_Upstream_Service_Time []string `json:"X-Envoy-Upstream-Service-Time"`
	X_Operation_Id []string  `json:"X-Operation-Id"`
}