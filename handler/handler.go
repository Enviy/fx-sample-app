package handler

import (
	"net/http"
	"sampleApp/controller"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Handlers .
type Handlers struct {
	logger *zap.Logger
	mux    *http.ServeMux
	ctlr   controller.Interface
	// Controller interface.
}

// Params .
type Params struct {
	fx.In

	Logger *zap.Logger
	Mux    *http.ServeMux
	Ctlr   controller.Interface
}

// New .
func New(p Params) *Handlers {
	h := &Handlers{
		logger: p.Logger,
		mux:    p.Mux,
		ctlr:   p.Ctlr,
	}
	h.RegisterHandlers()
	return h
}

// RegisterHandlers .
func (h *Handlers) RegisterHandlers() {
	h.mux.HandleFunc("/hello", h.Hello)
	h.mux.HandleFunc("/cat_fact", h.CatFact)
}

// Hello .
func (h *Handlers) Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))
	return
}

// CatFact .
func (h *Handlers) CatFact(w http.ResponseWriter, r *http.Request) {
	fact, err := h.ctlr.CatFact()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fact))
	return
}
