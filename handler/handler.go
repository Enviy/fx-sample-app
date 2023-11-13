package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"fx-sample-app/controller"

	"github.com/slack-go/slack"
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
	h.mux.HandleFunc("/ping", h.healthCheck)
	h.mux.HandleFunc("/hello", h.hello)
	h.mux.HandleFunc("/cat_fact", h.catFact)
	h.mux.HandleFunc("/cat_service", h.catsAAS)
}

// HealthCheck validate routing.
func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

// Hello .
func (h *Handlers) hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))
	return
}

// CatFact .
func (h *Handlers) catFact(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fact, err := h.con.CatFact(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fact))
	return
}

func (h *Handlers) catsAAS(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("ReadAll", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return the bytes to the body for the FormParser.
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Begin slack origin validation.
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")

	// Check if the request is within 5 minutes - replay attack
	now := time.Now()
	n, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		fmt.Printf("Err %d of type %T, %w\n", n, n, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if (now.Unix() - n) > 60*5 {
		fmt.Println("potential replay attack")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Build base signature string.
	sigBaseString := fmt.Sprintf("v0:%s:%s", timestamp, string(bodyBytes))

	// Collect app's signing key.
	// This would be moved to either config or secret store.
	signingKey := os.Getenv("SLACK_SIGNING_KEY")
	slackMac := r.Header.Get("X-Slack-Signature")

	// Validate Mac signatures are equal.
	if !validMAC([]byte(sigBaseString), []byte(slackMac), []byte(signingKey)) {
		fmt.Println("invalid signing key")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse request into SlashCommand object.
	slash, err := slack.SlashCommandParse(r)
	if err != nil {
		fmt.Println("SlashCommandParse", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate it's the correct slash command.
	if slash.Command != "/cat_fact" {
		fmt.Println("unsupported slash command", slash.Command)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Collect cat fact.
	catFact, err := h.con.CatFact(ctx)
	if err != nil {
		fmt.Println("controller.Cat", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Respond with cat fact.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(catFact))
	return
}

// validMAC reports whether messageMAC is a valid HMAC tag for message.
func validMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := "v0=" + string(mac.Sum(nil))
	return hmac.Equal(messageMAC, []byte(expectedMAC))
}
