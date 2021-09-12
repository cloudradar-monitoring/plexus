package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Error struct {
	Error       string
	Description string
}

func WriteBadRequestJSON(rw http.ResponseWriter, message string) {
	WriteJSONError(rw, http.StatusBadRequest, message)
}
func WriteBadGatewayJSON(rw http.ResponseWriter, message string) {
	WriteJSONError(rw, http.StatusBadGateway, message)
}

func WriteJSONError(rw http.ResponseWriter, code int, message string) {
	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(code)
	_ = json.NewEncoder(rw).Encode(err(code, message))
}

func WriteBadRequestText(rw http.ResponseWriter, message string) {
	WriteTextError(rw, http.StatusBadRequest, message)
}
func WriteBadGatewayText(rw http.ResponseWriter, message string) {
	WriteTextError(rw, http.StatusBadGateway, message)
}

func WriteTextError(rw http.ResponseWriter, code int, message string) {
	rw.Header().Add("content-type", "text/plain")
	rw.WriteHeader(code)
	fmt.Fprintln(rw, "Error:", http.StatusText(code))
	fmt.Fprintln(rw, "Desription:", message)
}

func err(code int, message string) *Error {
	return &Error{
		Error:       http.StatusText(code),
		Description: message,
	}
}
