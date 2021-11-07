package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"time"

	"golang.org/x/net/context/ctxhttp"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/token"
)

const (
	defaultTimeout = 5 * time.Second
	contentType    = "application/json"
)

var ErrUnableToPair = errors.New("unable to pair")

// CreateSession godoc
// @Summary Create a session.
// @Description Create a plexus session where meshagents can connect to.
// @Tags session
// @Accept application/x-www-form-urlencoded
// @Produce  application/json
// @Param id formData string true "session id"
// @Param ttl formData int true "the time to live for the session"
// @Param username formData string true "the credentials to open the remote control interface & delete the session"
// @Param password formData string true "the credentials to open the remote control interface & delete the session"
// @Param supporter_name formData string true "the supporter name"
// @Param supporter_avatar formData string true "the supporter avatar"
// @Success 200 {object} api.Session
// @Failure 400 {object} api.Error
// @Failure 500 {object} api.Error
// @Failure 502 {object} api.Error
// @Router /session [post]
func (h *Handler) CreateSession(rw http.ResponseWriter, r *http.Request) {
	if !h.auth(rw, r) {
		return
	}
	id := r.FormValue("id")
	ttlStr := r.FormValue("ttl")
	user := r.FormValue("username")
	pass := r.FormValue("password")
	supName := r.FormValue("supporter_name")
	supAvatar := r.FormValue("supporter_avatar")
	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("invalid ttl %s: %s", id, err))
		return
	}

	if !h.sessionCredentials && (user != "" || pass != "") {
		api.WriteBadRequestJSON(rw, "session credentials are not allowed")
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.sessions[id]; ok {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("session with id %s does already exist", id))
		return
	}

	mc, err := control.Connect(h.ccfg, h.log)
	defer mc.Close()
	if err != nil {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not connect to mesh control: %s", err))
		return
	}
	mesh, err := mc.CreateMesh(h.ccfg.MeshCentralGroupPrefix + "/" + id + "/" + token.New(5))
	if err != nil {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not create mesh: %s", err))
		return
	}
	serverInfo, err := mc.ServerInfo()
	if err != nil {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not get server id: %s", err))
		return
	}
	sessionToken := token.NewAuth()

	session := &Session{
		Token:     sessionToken,
		ID:        id,
		Username:  user,
		Password:  pass,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
		AgentConfig: api.AgentConfig{
			MeshName:   mesh.Name,
			MeshIDHex:  mesh.IDHex,
			MeshID:     mesh.ID,
			ServerID:   serverInfo.AgentHash,
			MeshServer: fmt.Sprintf("wss://%s%s", r.Host, path.Join(h.prefix, "agent", fmt.Sprintf("%s:%s", id, sessionToken))),
			MeshType:   2,
		},
	}

	if h.pcfg.PairingURL != "" {
		h.log.Debugf("pairing to %s ...", h.pcfg.PairingURL)

		if supName == "" || supAvatar == "" {
			api.WriteJSONError(rw, http.StatusBadRequest, "You need to provide supporter_name and supporter_avatar for pairing")
			return
		}

		pr, err := h.pcPair(r.Context(), h.pcfg.PairingURL, &Request{
			Url: fmt.Sprintf("https://%s%s/pairing", h.pcfg.ServerAddress, h.prefix),
		})
		if err != nil {
			api.WriteJSONError(rw, http.StatusBadGateway, fmt.Sprintf("Unable to create session, failed to pair: %s", err.Error()))
			return
		}

		h.log.Debugf("pairing succeeded code(%s)", pr.Code)

		session.PairingCode = pr.Code
		session.PairingURL = pr.PairingURL
		session.SupporterName = supName
		session.SupporterAvatar = supAvatar
	}

	h.sessions[id] = session

	go func() {
		<-time.After(time.Duration(ttl) * time.Second)
		h.lock.Lock()
		defer h.lock.Unlock()
		if s, ok := h.sessions[id]; ok {
			h.log.Infof("Session %s expired", id)
			err := h.deleteInternal(s)
			if err != nil {
				h.log.Errorf("Could not clean session %s: %s", id, err)
			}
		}
	}()

	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(rw).Encode(&api.Session{
		ID:          id,
		SessionURL:  fmt.Sprintf("https://%s%s", r.Host, path.Join(h.prefix, "session", id)),
		AgentMSH:    fmt.Sprintf("https://%s%s", r.Host, path.Join(h.prefix, "config", fmt.Sprintf("%s:%s", id, sessionToken))),
		PairingCode: session.PairingCode,
		PairingURL:  session.PairingURL,
		AgentConfig: session.AgentConfig,
		ExpiresAt:   session.ExpiresAt,
	})
}

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
	jsonRequest, _ := json.Marshal(req)
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
		h.log.Errorf("pairing request failed: status(%d) response(%s)", response.StatusCode, string(jsonResponse))
		return nil, fmt.Errorf("reading body failed code(%d) error: %w", response.StatusCode, err)
	}

	if response.StatusCode != http.StatusOK {
		h.log.Errorf("pairing request failed: status(%d) response(%s)", response.StatusCode, string(jsonResponse))
		return nil, fmt.Errorf("code(%d) error: %w", response.StatusCode, ErrUnableToPair)
	}

	resp := Response{}
	err = json.Unmarshal(jsonResponse, &resp)
	if err != nil {
		h.log.Errorf("pairing request failed: status(%d) response(%s)", response.StatusCode, string(jsonResponse))
		return nil, fmt.Errorf("unmarshaling response failed code(%d) error: %w", response.StatusCode, err)
	}

	return &resp, nil
}
