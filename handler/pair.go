package handler

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"

	"github.com/cloudradar-monitoring/plexus/api"
)

// Pair godoc
// @Summary Gets the pairing code info.
// @Tags session
// @Produce application/json
// @Param code path string true "code"
// @Security BasicAuth
// @Success 200 {object} api.URLResponse
// @Failure 401 {object} api.Error
// @Failure 404 {object} api.Error
// @Router /pair/{code} [get]
func (h *Handler) Pair(rw http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	defer h.lock.Unlock()
	id, ok := h.codes[mux.Vars(r)["id"]]
	if !ok {
		api.WriteJSONError(rw, http.StatusNotFound, "Code not found")
		return
	}

	session, ok := h.checkSessionAuthentication(rw, r, id)
	if !ok {
		return
	}

	api.WriteJSONResponse(rw, http.StatusNotFound, api.PairedSession{
		AgentMSH:        fmt.Sprintf("https://%s%s", r.Host, path.Join(h.prefix, "config", fmt.Sprintf("%s:%s", id, session.Token))),
		SupporterName:   session.SupporterName,
		SupporterAvatar: session.SupporterAvatar,
		CompanyName:     h.pcfg.CompanyName,
		CompanyLogo:     h.pcfg.CompanyLogo,
		ExpiresAt:       session.ExpiresAt,
	})
}
