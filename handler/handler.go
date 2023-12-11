package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"fx-sample-app/controller"
	pb "fx-sample-app/proto"

	"github.com/slack-go/slack"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Handlers implements grpc service.
type Handlers struct {
	pb.UnimplementedSampleServer

	log *zap.Logger
	con controller.Controller
}

// Params defines constructor requirements.
type Params struct {
	fx.In

	Log *zap.Logger
	Lc  *fx.Lifecycle
	Cfg config.Provider
	Con controller.Controller
}

// New is the handler constructor.
func New(p Params) (*Handlers, error) {
	h := &Handlers{
		log: p.Log,
		mux: p.Mux,
		con: p.Con,
	}
	ln, err := net.Listen(
		"tcp",
		p.Cfg.Get("server.address").String(),
	)
	if err != nil {
		return nil, err
	}

	// Create grpc server, register service.
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	pb.RegisterSampleServer(grpcServer, h)

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			h.log.Info("Serving gRPC",
				zap.String("address", p.Cfg.Get("server.address").String()),
			)
			go func() {
				err := grpcServer.Serve(ln)
				if err != nil {
					h.log.Error("grpc failed to serve",
						zap.Error(err),
					)
					return
				}
			}()
		},
		OnStop: func(ctx context.Context) error {
			h.log.Info("shutting down")
		},
	})

	return h, nil
}

// Hello .
func (h *Handlers) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Greeting: "Hello " + req.Name,
	}, nil
}

// CatFact returns a random cat fact.
func (h *Handlers) CatFact(ctx context.Context, req *pb.CatFactRequest) (*pb.CatFactResponse, error) {
	fact, err := h.con.CatFact(ctx)
	if err != nil {
		return &pb.CatFactResponse{}, err
	}

	return &pb.CatFactResponse{
		Fact: fact,
	}, nil
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
