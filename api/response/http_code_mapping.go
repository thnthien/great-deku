package response

import (
	"encoding/json"
	"net/http"

	"github.com/thnthien/great-deku/errors"
)

var HttpStatusMap = map[errors.Code]int{}

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
