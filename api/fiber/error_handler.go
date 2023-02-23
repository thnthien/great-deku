package fiberapi

import (
	"github.com/thnthien/great-deku/api/response"
	"unsafe"

	"github.com/gofiber/fiber/v2"

	"github.com/thnthien/great-deku/api"
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
				Status:     errors.Internal.String(),
				Code:       uint32(errors.Internal),
				Message:    errors.Internal.String(),
				DevMessage: devMsg,
				Errors:     nil,
				RID:        rid,
			})
		}

		clientError, ok := err.(*errors.APIError)
		if !ok {
			return ctx.Status(code).JSON(api.HTTPErrorResponse{
				Status:     errors.Internal.String(),
				Code:       uint32(errors.Internal),
				Message:    err.Error(),
				DevMessage: devMsg,
				Errors:     nil,
				RID:        rid,
			})
		}

		devMsg = clientError
		if env != "D" {
			devMsg = nil
		}
		code = mappingFunc(clientError.Code)
		resStatus := clientError.Code.String()
		resCode := uint32(clientError.Code)
		if clientError.XCode != errors.OK {
			resStatus = clientError.XCode.String()
			resCode = uint32(clientError.XCode)
		}

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
