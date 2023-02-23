package response

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

// IResponse ...
type IResponse interface {
	WithData(data interface{}) IResponse
	WithPaging(data interface{}) IResponse
	WithMessage(data string) IResponse
	WithStatus(status int) IResponse
	Json(c *fiber.Ctx) error
	NoContent(c *fiber.Ctx) error
}

type cursors struct {
	After  string `json:"after"`
	Before string `json:"before"`
}

type links struct {
	BeforeCount int      `json:"before_count,omitempty"`
	AfterCount  int      `json:"after_count,omitempty"`
	Count       int      `json:"count,omitempty"`
	Cursors     *cursors `json:"cursors,omitempty"`
	Next        string   `json:"next"`
	Prev        string   `json:"prev"`
}

type commonFormatResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Link    *links      `json:"link,omitempty"`
	Message string      `json:"message,omitempty"`
}

type response struct {
	status  int
	data    interface{}
	link    *links
	message string
}

// NewResponse ...
func NewResponse() IResponse {
	return response{}
}

// WithData ...
func (r response) WithData(data interface{}) IResponse {
	r.data = data
	return r
}

// WithMessage ...
func (r response) WithMessage(data string) IResponse {
	r.message = data
	return r
}

// WithPaging ...
func (r response) WithPaging(data interface{}) IResponse {
	if r.link == nil {
		r.link = &links{}
	}
	err := copier.Copy(r.link, data)
	if err != nil {
		fmt.Printf("[WithPaging] failed to copy paging data [%+v] [%v]", data, err)
	}
	return r
}

// WithStatus ...
func (r response) WithStatus(status int) IResponse {
	r.status = status
	return r
}

// Json ...
func (r response) Json(c *fiber.Ctx) error {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return c.Status(r.status).JSON(commonFormatResponse{
		Data:    r.data,
		Link:    r.link,
		Message: r.message,
	})
}

// NoContent ...
func (r response) NoContent(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

var _ IResponse = response{}
