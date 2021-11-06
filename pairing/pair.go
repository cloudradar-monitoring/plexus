package pairing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
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

func Pair(ctx context.Context, url string, req *Request) (*Response, error) {
	jsonRequest, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: defaultTimeout,
	}

	response, err := ctxhttp.Post(ctx, client, url, contentType, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		log.Error().Msg(string(jsonResponse))
		return nil, ErrUnableToPair
	}

	resp := Response{}
	err = json.Unmarshal(jsonResponse, &resp)
	if err != nil {
		return nil, err
	}



	return &resp, nil
}
