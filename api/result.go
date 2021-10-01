package api

import (
	"encoding/json"
	"net/http"
)

type Result struct {
	Result string
}

func WriteResult(rw http.ResponseWriter, code int, result string) {
	rw.Header().Add("content-type", "text/plain")
	rw.WriteHeader(code)
	_ = json.NewEncoder(rw).Encode(&Result{Result: result})
}

func WriteJSONResponse(rw http.ResponseWriter, code int, response interface{}) {
	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(code)
	_ = json.NewEncoder(rw).Encode(response)
}

type URLResponse struct {
	URL string `json:"url"`
}
