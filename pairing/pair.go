package pairing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

const defaultTimeout = 5 * time.Second

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

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Timeout: defaultTimeout,
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, ErrUnableToPair
	}

	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	resp := Response{}
	err = json.Unmarshal(jsonResponse, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
