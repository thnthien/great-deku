package api

import (
	"github.com/thnthien/great-deku/api/response"
)

// HTTPErrorResponse ...
type HTTPErrorResponse struct {
	Status     string                 `json:"status"`
	Code       uint32                 `json:"code"`
	Message    string                 `json:"message"`
	DevMessage interface{}            `json:"dev_message" swaggerignore:"true"`
	Errors     map[string]interface{} `json:"errors"`
	RID        string                 `json:"rid"`
}

// Handler ...
type Handler struct {
}

// Resp ...
func (Handler) Resp() response.IResponse {
	return response.NewResponse()
}

type IHealthController interface {
	SetReady(b bool)
}
