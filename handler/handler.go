package handler

import (
	"net/http"

	"fx-sample-app/controller"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Handlers .
type Handlers struct {
	log *zap.Logger
	mux *http.ServeMux
	con controller.Controller
}

// Params .
type Params struct {
	fx.In

	Log *zap.Logger
	Mux *http.ServeMux
	Con controller.Controller
}

// New .
func New(p Params) *Handlers {
	h := &Handlers{
		log: p.Log,
		mux: p.Mux,
		con: p.Con,
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
	fact, err := h.con.CatFact()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fact))
	return
}
