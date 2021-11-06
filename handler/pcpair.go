package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context/ctxhttp"
)

const (
	defaultTimeout = 5 * time.Second
	contentType    = "application/json"
)

var ErrUnableToPair = errors.New("unable to pair")

type Request struct {
	Url string `json:"url"`
}

type Response struct {
	Success     bool   `json:"success"`
	Code        string `json:"code"`
	PairingURL  string `json:"pairing_url"`
	RedirectURL string `json:"redirect_url"`
}

func (h *Handler) pcPair(ctx context.Context, url string, req *Request) (*Response, error) {
	jsonRequest, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request failed: %w", err)
	}
	client := &http.Client{
		Timeout: defaultTimeout,
	}

	response, err := ctxhttp.Post(ctx, client, url, contentType, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, fmt.Errorf("post failed: %w", err)
	}

	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body failed: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		h.log.Errorf("pairing request failed: status(%d) response(%s)", response.StatusCode, string(jsonResponse))
		return nil, ErrUnableToPair
	}

	resp := Response{}
	err = json.Unmarshal(jsonResponse, &resp)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling response failed: %w", err)
	}

	return &resp, nil
}
