package response

import (
	"encoding/json"
	"net/http"

	"github.com/thnthien/great-deku/errors"
)

var HttpStatusMap = map[errors.Code]int{
	errors.OK:                 http.StatusOK,
	errors.Canceled:           http.StatusRequestTimeout,
	errors.Unknown:            http.StatusInternalServerError,
	errors.Internal:           http.StatusInternalServerError,
	errors.DataLoss:           http.StatusInternalServerError,
	errors.InvalidArgument:    http.StatusBadRequest,
	errors.OutOfRange:         http.StatusBadRequest,
	errors.DeadlineExceeded:   http.StatusGatewayTimeout,
	errors.NotFound:           http.StatusNotFound,
	errors.AlreadyExists:      http.StatusConflict,
	errors.Aborted:            http.StatusConflict,
	errors.PermissionDenied:   http.StatusForbidden,
	errors.Unauthenticated:    http.StatusUnauthorized,
	errors.ResourceExhausted:  http.StatusTooManyRequests,
	errors.FailedPrecondition: http.StatusPreconditionFailed,
	errors.Unimplemented:      http.StatusNotImplemented,
	errors.Unavailable:        http.StatusServiceUnavailable,
}

func DefaultStatusMapping(code errors.Code) int {
	status, ok := HttpStatusMap[code]
	if ok {
		return status
	}
	return http.StatusInternalServerError
}

func Write(w http.ResponseWriter, data string) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Json(w http.ResponseWriter, data interface{}) {
	responseJSON(w, data)
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func JsonError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

}
