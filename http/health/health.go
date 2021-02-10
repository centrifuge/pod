package health

import (
	"net/http"

	"github.com/centrifuge/go-centrifuge/version"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// config defines required methods for Health handler
type config interface {
	GetNetworkString() string
}

// handler handles the Health APIs
type handler struct {
	c config
}

// Pong is the response for Ping
type Pong struct {
	Version string `json:"version"`
	Network string `json:"network"`
}

// Ping responds with node version and network

// @summary Health check for Node
// @description returns node version and network
// @id ping
// @tags Health
// @produce json
// @success 200 {object} health.Pong
// @router /ping [get]
func (h handler) Ping(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, Pong{
		Version: version.GetVersion().String(),
		Network: h.c.GetNetworkString(),
	})
}

// Register registers the health APIs to the router
func Register(r chi.Router, config config) {
	h := handler{c: config}
	r.Get("/ping", h.Ping)
}
