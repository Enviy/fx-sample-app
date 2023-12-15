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
	"strings"
	"time"

	"fx-sample-app/controller"
	pb "fx-sample-app/proto/fxsample"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/slack-go/slack"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Handlers implements grpc service.
type Handlers struct {
	pb.UnimplementedFxsampleServer

	log    *zap.Logger
	con    controller.Controller
	health *health.Server
}

// Params defines constructor requirements.
type Params struct {
	fx.In

	Log *zap.Logger
	Lc  fx.Lifecycle
	Cfg config.Provider
	Con controller.Controller
}

// New is the handler constructor.
func New(p Params) (*Handlers, error) {
	h := &Handlers{
		log: p.Log,
		con: p.Con,
	}
	ln, err := net.Listen(
		"tcp",
		p.Cfg.Get("server.address").String(),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc net listen %w", err)
	}

	// Create grpc server.
	grpcServer := grpc.NewServer()

	// Add reflection to service stack.
	reflection.Register(grpcServer)

	// Add healthcheck to service stack.
	healthCheck := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthCheck)
	h.health = healthCheck

	// Add sample proto service to service stack.
	pb.RegisterFxsampleServer(grpcServer, h)

	// gRPC client connection for HTTP proxy.
	conn, err := grpc.DialContext(
		context.Background(),
		p.Cfg.Get("server.address").String(),
		//grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial context %w", err)
	}

	// Define header forwarders.
	timeHeader := runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		if strings.EqualFold(key, "X-Slack-Request-Timestamp") {
			return key, true
		}
		return "", false
	})
	signHeader := runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		if strings.EqualFold(key, "X-Slack-Signature") {
			return key, true
		}
		return "", false
	})

	// Create proxy mux.
	gwmux := runtime.NewServeMux(timeHeader, signHeader)

	// Register proxy handlers. Routes http calls to gRPC.
	err = pb.RegisterFxsampleHandler(
		context.Background(),
		gwmux,
		conn,
	)
	if err != nil {
		return nil, fmt.Errorf("register proxy handler %w", err)
	}

	gwServer := &http.Server{
		Addr:    "127.0.0.1:8090",
		Handler: gwmux,
	}

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start gRPC server.
			h.log.Info("Serving gRPC",
				zap.String("address", p.Cfg.Get("server.address").String()),
			)
			go func() {
				if err := grpcServer.Serve(ln); err != nil {
					h.log.Error("grpc serve", zap.Error(err))
					return
				}
			}()

			// Start proxy server.
			h.log.Info("Starting http proxy", zap.String("address", "127.0.0.1:8090"))
			go func() {
				if err := gwServer.ListenAndServe(); err != nil {
					h.log.Error("proxy listen&serve", zap.Error(err))
					return
				}
			}()

			// Set initial health status.
			h.health.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			h.log.Info("shutting down")
			grpcServer.GracefulStop()
			gwServer.Shutdown(ctx)
			return nil
		},
	})

	return h, nil
}

// Hello .
func (h *Handlers) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "CatFact: missing incoming metadata in rpc context")
	}

	if v, ok := md["X-Slack-Request-Timestamp"]; ok {
		h.log.Info("timestamp header", zap.String("timestamp", v))
	}
	if v, ok := md["X-Slack-Signature"]; ok {
		h.log.Info("signature header", zap.String("signature", v))
	}

	body := req.GetBodyBytes().GetData()
	h.log.Info("body bytes", zap.String("body", string(body)))

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
