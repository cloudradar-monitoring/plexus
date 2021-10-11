package handler

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/asset"
)

func (h *Handler) Home(rw http.ResponseWriter, r *http.Request) {
	if !h.auth(rw, r) {
		return
	}

	rw.Header().Set("content-type", "text/html")
	if _, err := rw.Write(asset.IndexHTML); err != nil {
		log.Error().Err(err).Msg("unable to execute index.html template")
		api.WriteJSONError(rw, http.StatusInternalServerError, "Unable to process the request, try again later.")
	}
}
