package fiberapi

import (
	"unsafe"

	"github.com/gofiber/fiber/v2"

	"github.com/thnthien/great-deku/api"
	"github.com/thnthien/great-deku/api/response"
	"github.com/thnthien/great-deku/errors"
)

func getString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func ErrorHandler(env string, httpStatusMappingFunc func(code errors.Code) int) func(ctx *fiber.Ctx, err error) error {
	mappingFunc := httpStatusMappingFunc
	if mappingFunc == nil {
		mappingFunc = response.DefaultStatusMapping
	}
	return func(ctx *fiber.Ctx, err error) error {
		// Statuscode defaults to 500
		code := fiber.StatusInternalServerError

		rid := getString(ctx.Context().Response.Header.Peek(fiber.HeaderXRequestID))

		devMsg := err
		if env != "D" {
			devMsg = nil
		}

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			return ctx.Status(code).JSON(api.HTTPErrorResponse{
				Status:     "Internal Server Error",
				Code:       500,
				Message:    "Internal Server Error",
				DevMessage: devMsg,
				Errors:     nil,
				RID:        rid,
			})
		}

		clientError, ok := err.(*errors.APIError)
		if !ok {
			return ctx.Status(code).JSON(api.HTTPErrorResponse{
				Status:     "Internal Server Error",
				Code:       500,
				Message:    "Internal Server Error",
				DevMessage: devMsg,
				Errors:     nil,
				RID:        rid,
			})
		}

		devMsg = clientError
		if env != "D" {
			devMsg = nil
		}
		xcode := clientError.Code
		code = mappingFunc(xcode)
		resStatus := xcode.String()
		resCode := uint32(clientError.Code)

		return ctx.Status(code).JSON(api.HTTPErrorResponse{
			Status:     resStatus,
			Code:       resCode,
			Message:    clientError.Message,
			DevMessage: devMsg,
			Errors:     nil,
			RID:        rid,
		})
	}
}
